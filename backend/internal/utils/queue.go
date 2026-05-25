package utils

import (
	"time"

	"github.com/nats-io/nats.go"
)

// QueueLagMs calculates the lag between message timestamp and current time in milliseconds.
func QueueLagMs(msg *nats.Msg) int64 {
	meta, err := msg.Metadata()
	if err != nil {
		return 0
	}
	return time.Since(meta.Timestamp).Milliseconds()
}
