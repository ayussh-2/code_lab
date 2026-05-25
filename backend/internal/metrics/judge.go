package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type JudgeMetrics struct {
	received atomic.Uint64
	errors   atomic.Uint64
	verdicts sync.Map
}

func NewJudgeMetrics() *JudgeMetrics {
	return &JudgeMetrics{}
}

func (m *JudgeMetrics) IncReceived() {
	m.received.Add(1)
}

func (m *JudgeMetrics) IncErrors() {
	m.errors.Add(1)
}

func (m *JudgeMetrics) IncVerdict(verdict string) {
	value, _ := m.verdicts.LoadOrStore(verdict, &atomic.Uint64{})
	value.(*atomic.Uint64).Add(1)
}

func (m *JudgeMetrics) LogEveryMinute(ctx context.Context, log *zap.Logger) {
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
