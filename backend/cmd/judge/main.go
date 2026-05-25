package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ayussh-2/config"
	"github.com/ayussh-2/internal/worker/judge"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	// Validate configuration
	if err := judge.ValidateConfig(cfg); err != nil {
		panic("invalid configuration: " + err.Error())
	}

	// Initialize all dependencies
	deps, err := judge.Bootstrap(cfg)
	if err != nil {
		panic("failed to bootstrap: " + err.Error())
	}
	defer deps.Close()

	log := deps.Log
	svc := deps.Service()
	handler := judge.NewMessageHandler(log, svc, deps.Metrics)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start metrics logging
	deps.Metrics.LogEveryMinute(ctx, log)

	// Subscribe to judge subject
	sub, err := deps.Queue.Subscribe(ctx, cfg.JudgeSubject, "judge-worker", cfg.JudgeConcurrency, handler.Handle)
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
