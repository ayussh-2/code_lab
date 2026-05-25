package judge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ayussh-2/internal/metrics"
	"github.com/ayussh-2/internal/services"
	"github.com/ayussh-2/internal/utils"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// MessageHandler processes judge messages from the queue.
type MessageHandler struct {
	log     *zap.Logger
	service *services.SubmissionService
	metrics *metrics.JudgeMetrics
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(log *zap.Logger, service *services.SubmissionService, metrics *metrics.JudgeMetrics) *MessageHandler {
	return &MessageHandler{
		log:     log,
		service: service,
		metrics: metrics,
	}
}

// Handle processes a single judge message.
func (h *MessageHandler) Handle(ctx context.Context, msg *nats.Msg) error {
	var payload services.JudgeMessage
	h.metrics.IncReceived()

	defer func() {
		if recovered := recover(); recovered != nil {
			message := fmt.Sprintf("judge panic: %v", recovered)
			h.metrics.IncErrors()
			h.log.Error("Judge handler panicked", zap.Any("panic", recovered), zap.Uint("submission_id", payload.SubmissionID))
			if payload.SubmissionID != 0 {
				if markErr := h.service.MarkInternalError(payload.SubmissionID, message); markErr != nil {
					h.log.Error("Failed to mark panic as internal error", zap.Error(markErr), zap.Uint("submission_id", payload.SubmissionID))
				}
			}
			h.metrics.IncVerdict("IE")
		}
	}()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		h.metrics.IncErrors()
		h.log.Error("Invalid judge message", zap.Error(err))
		return nil
	}

	if payload.SubmissionID == 0 {
		h.metrics.IncErrors()
		h.log.Error("Invalid judge message: missing submission_id")
		return nil
	}

	return h.judgeSubmission(ctx, payload, msg)
}

func (h *MessageHandler) judgeSubmission(ctx context.Context, payload services.JudgeMessage, msg *nats.Msg) error {
	runLog := h.log.With(
		zap.Uint("submission_id", payload.SubmissionID),
		zap.Int64("queue_lag_ms", utils.QueueLagMs(msg)),
	)

	if err := h.service.Judge(payload.SubmissionID); err != nil {
		h.metrics.IncErrors()
		runLog.Error("Judge failed; message will be retried", zap.Error(err))
		return err
	}

	verdict, err := h.service.Verdict(payload.SubmissionID)
	if err != nil {
		h.metrics.IncErrors()
		runLog.Error("Judge verdict lookup failed", zap.Error(err))
		return err
	}

	h.metrics.IncVerdict(verdict)
	runLog.Info("Judge completed", zap.String("verdict", verdict))
	return nil
}
