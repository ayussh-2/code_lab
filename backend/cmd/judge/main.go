package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/database"
	"github.com/ayussh-2/internal/logger"
	"github.com/ayussh-2/internal/queue"
	"github.com/ayussh-2/internal/sandbox/docker"
	"github.com/ayussh-2/internal/services"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	log, err := logger.Init(cfg.Env)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("Cannot connect to DB", zap.Error(err))
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Cannot access DB handle", zap.Error(err))
	}
	defer sqlDB.Close()

	queueClient, err := queue.New(cfg)
	if err != nil {
		log.Fatal("Cannot connect to NATS", zap.Error(err))
	}
	defer queueClient.Close()

	if err := queueClient.EnsureStream(cfg.JudgeStream, []string{cfg.JudgeSubject}); err != nil {
		log.Fatal("Cannot ensure judge stream", zap.Error(err))
	}

	dockerCli, err := docker.NewDefaultClient()
	if err != nil {
		log.Fatal("Cannot connect to docker", zap.Error(err))
	}
	defer dockerCli.Close()

	runner := docker.NewRunner(dockerCli, cfg.SandboxWorkDir)
	svc := services.NewSubmissionService(log, db, cfg, runner, nil)
	metrics := newJudgeMetrics()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	metrics.logEveryMinute(ctx, log)

	sub, err := queueClient.Subscribe(ctx, cfg.JudgeSubject, "judge-worker", cfg.JudgeConcurrency, func(ctx context.Context, msg *nats.Msg) error {
		return handleJudgeMessage(ctx, log, svc, metrics, msg)
	})
	if err != nil {
		log.Fatal("Cannot subscribe to judge subject", zap.Error(err))
	}

	log.Info("Judge worker started",
		zap.String("stream", cfg.JudgeStream),
		zap.String("subject", cfg.JudgeSubject),
		zap.Int("concurrency", cfg.JudgeConcurrency),
	)

	<-ctx.Done()
	log.Info("Judge worker shutting down")
	if err := sub.Stop(); err != nil {
		log.Error("Judge subscription shutdown failed", zap.Error(err))
	}
}

func handleJudgeMessage(ctx context.Context, log *zap.Logger, svc *services.SubmissionService, metrics *judgeMetrics, msg *nats.Msg) (err error) {
	var payload services.JudgeMessage
	metrics.received.Add(1)

	defer func() {
		if recovered := recover(); recovered != nil {
			message := fmt.Sprintf("judge panic: %v", recovered)
			metrics.errors.Add(1)
			log.Error("Judge handler panicked", zap.Any("panic", recovered), zap.Uint("submission_id", payload.SubmissionID))
			if payload.SubmissionID != 0 {
				if markErr := svc.MarkInternalError(payload.SubmissionID, message); markErr != nil {
					log.Error("Failed to mark panic as internal error", zap.Error(markErr), zap.Uint("submission_id", payload.SubmissionID))
					err = markErr
					return
				}
			}
			metrics.incVerdict("IE")
			err = nil
		}
	}()

	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		metrics.errors.Add(1)
		log.Error("Invalid judge message", zap.Error(err))
		return nil
	}
	if payload.SubmissionID == 0 {
		metrics.errors.Add(1)
		log.Error("Invalid judge message: missing submission_id")
		return nil
	}

	runLog := log.With(
		zap.Uint("submission_id", payload.SubmissionID),
		zap.Int64("queue_lag_ms", queueLagMs(msg)),
	)
	if err := svc.Judge(payload.SubmissionID); err != nil {
		metrics.errors.Add(1)
		runLog.Error("Judge failed; message will be retried", zap.Error(err))
		return err
	}

	verdict, err := svc.Verdict(payload.SubmissionID)
	if err != nil {
		metrics.errors.Add(1)
		runLog.Error("Judge verdict lookup failed", zap.Error(err))
		return err
	}
	metrics.incVerdict(verdict)
	runLog.Info("Judge completed", zap.String("verdict", verdict))
	return nil
}

type judgeMetrics struct {
	received atomic.Uint64
	errors   atomic.Uint64
	verdicts sync.Map
}

func newJudgeMetrics() *judgeMetrics {
	return &judgeMetrics{}
}

func (m *judgeMetrics) incVerdict(verdict string) {
	value, _ := m.verdicts.LoadOrStore(verdict, &atomic.Uint64{})
	value.(*atomic.Uint64).Add(1)
}

func (m *judgeMetrics) logEveryMinute(ctx context.Context, log *zap.Logger) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fields := []zap.Field{
					zap.Uint64("submissions_received_total", m.received.Load()),
					zap.Uint64("judge_errors_total", m.errors.Load()),
				}
				m.verdicts.Range(func(key, value any) bool {
					fields = append(fields, zap.Uint64("verdicts_total_"+key.(string), value.(*atomic.Uint64).Load()))
					return true
				})
				log.Info("Judge metrics", fields...)
			}
		}
	}()
}

func queueLagMs(msg *nats.Msg) int64 {
	meta, err := msg.Metadata()
	if err != nil {
		return 0
	}
	return time.Since(meta.Timestamp).Milliseconds()
}
