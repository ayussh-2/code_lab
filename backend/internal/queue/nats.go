package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ayussh-2/config"
	"github.com/nats-io/nats.go"
)

const nackDelay = 5 * time.Second

type Client struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

type Subscription struct {
	sub    *nats.Subscription
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(cfg *config.Config) (*Client, error) {
	conn, err := nats.Connect(cfg.NATSURL, nats.Name("codelab-judge"))
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client{conn: conn, js: js}, nil
}

func (c *Client) EnsureStream(name string, subjects []string) error {
	info, err := c.js.StreamInfo(name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = c.js.AddStream(&nats.StreamConfig{
			Name:     name,
			Subjects: subjects,
			Storage:  nats.FileStorage,
		})
		return err
	}
	if err != nil {
		return err
	}
	if sameSubjects(info.Config.Subjects, subjects) {
		return nil
	}

	cfg := info.Config
	cfg.Subjects = subjects
	_, err = c.js.UpdateStream(&cfg)
	return err
}

func (c *Client) Publish(subject string, payload []byte) error {
	_, err := c.js.Publish(subject, payload)
	return err
}

func (c *Client) Subscribe(
	ctx context.Context,
	subject string,
	durable string,
	concurrency int,
	handler func(context.Context, *nats.Msg) error,
) (*Subscription, error) {
	if concurrency <= 0 {
		concurrency = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	sem := make(chan struct{}, concurrency)
	subscription := &Subscription{cancel: cancel}

	sub, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		select {
		case <-ctx.Done():
			_ = msg.NakWithDelay(nackDelay)
			return
		case sem <- struct{}{}:
		}

		subscription.wg.Add(1)
		go func() {
			defer subscription.wg.Done()
			defer func() { <-sem }()

			if err := handler(ctx, msg); err != nil {
				_ = msg.NakWithDelay(nackDelay)
				return
			}
			_ = msg.Ack()
		}()
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.MaxAckPending(concurrency),
	)
	if err != nil {
		cancel()
		return nil, err
	}

	subscription.sub = sub
	return subscription, nil
}

func (s *Subscription) Stop() error {
	s.cancel()
	err := s.sub.Drain()
	s.wg.Wait()
	return err
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

func sameSubjects(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := make(map[string]struct{}, len(a))
	for _, subject := range a {
		seen[subject] = struct{}{}
	}
	for _, subject := range b {
		if _, ok := seen[subject]; !ok {
			return false
		}
	}
	return true
}
