package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSQueue implements Queue with JetStream durable consumers.
type NATSQueue struct {
	js  nats.JetStreamContext
	cfg Config
}

// NewNATSQueue provisions stream/consumer and returns a queue implementation.
func NewNATSQueue(nc *nats.Conn, cfg Config) (*NATSQueue, error) {
	if nc == nil {
		return nil, errors.New("queue: nats connection is nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("queue: jetstream: %w", err)
	}
	if err := ensureStream(js, cfg); err != nil {
		return nil, err
	}
	if err := ensureConsumer(js, cfg); err != nil {
		return nil, err
	}
	return &NATSQueue{js: js, cfg: cfg}, nil
}

func ensureStream(js nats.JetStreamContext, cfg Config) error {
	info, err := js.StreamInfo(cfg.StreamName)
	if err == nil && info != nil {
		return nil
	}
	if err != nil && !errors.Is(err, nats.ErrStreamNotFound) {
		return fmt.Errorf("queue: stream info: %w", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      cfg.StreamName,
		Subjects:  []string{cfg.Subject, cfg.DeadLetter},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
	})
	if err != nil {
		return fmt.Errorf("queue: add stream: %w", err)
	}
	return nil
}

func ensureConsumer(js nats.JetStreamContext, cfg Config) error {
	_, err := js.ConsumerInfo(cfg.StreamName, cfg.ConsumerName)
	if err == nil {
		return nil
	}
	if !errors.Is(err, nats.ErrConsumerNotFound) {
		return fmt.Errorf("queue: consumer info: %w", err)
	}
	_, err = js.AddConsumer(cfg.StreamName, &nats.ConsumerConfig{
		Durable:       cfg.ConsumerName,
		FilterSubject: cfg.Subject,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       cfg.RetryDelay(cfg.MaxRetries+1) + time.Second,
		MaxDeliver:    -1, // application owns dead-letter decision.
	})
	if err != nil {
		return fmt.Errorf("queue: add consumer: %w", err)
	}
	return nil
}

// Enqueue publishes one job message to the main work subject.
func (q *NATSQueue) Enqueue(ctx context.Context, m Message) error {
	if m.JobID == "" {
		return errors.New("queue: job_id is required")
	}
	body, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("queue: marshal message: %w", err)
	}
	msg := nats.NewMsg(q.cfg.Subject)
	msg.Data = body
	if _, err := q.js.PublishMsg(msg, nats.Context(ctx)); err != nil {
		return fmt.Errorf("queue: enqueue publish: %w", err)
	}
	return nil
}

// Consume processes messages until ctx is canceled.
func (q *NATSQueue) Consume(ctx context.Context, h Handler) error {
	if h == nil {
		return errors.New("queue: handler is nil")
	}
	sub, err := q.js.PullSubscribe(q.cfg.Subject, q.cfg.ConsumerName, nats.BindStream(q.cfg.StreamName))
	if err != nil {
		return fmt.Errorf("queue: subscribe: %w", err)
	}
	defer func() { _ = sub.Unsubscribe() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msgs, err := sub.Fetch(1, nats.MaxWait(200*time.Millisecond))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue
			}
			return fmt.Errorf("queue: fetch: %w", err)
		}
		for _, msg := range msgs {
			if err := q.handleOne(ctx, h, msg); err != nil {
				return err
			}
		}
	}
}

func (q *NATSQueue) handleOne(ctx context.Context, h Handler, msg *nats.Msg) error {
	var m Message
	if err := json.Unmarshal(msg.Data, &m); err != nil {
		if ackErr := msg.Ack(); ackErr != nil {
			return fmt.Errorf("queue: ack invalid payload: %w", ackErr)
		}
		return nil
	}
	meta, err := msg.Metadata()
	if err != nil {
		return fmt.Errorf("queue: metadata: %w", err)
	}
	attempt := 1
	if meta.NumDelivered > 0 {
		if meta.NumDelivered > uint64(math.MaxInt) {
			attempt = math.MaxInt
		} else {
			attempt = int(meta.NumDelivered)
		}
	}
	if attempt < 1 {
		attempt = 1
	}

	if err := h(ctx, m); err == nil {
		if ackErr := msg.Ack(); ackErr != nil {
			return fmt.Errorf("queue: ack success: %w", ackErr)
		}
		return nil
	}

	if attempt > q.cfg.MaxRetries {
		if err := q.publishDLQ(ctx, m, attempt); err != nil {
			return err
		}
		if ackErr := msg.Ack(); ackErr != nil {
			return fmt.Errorf("queue: ack dead-lettered: %w", ackErr)
		}
		return nil
	}

	if nakErr := msg.NakWithDelay(q.cfg.RetryDelay(attempt)); nakErr != nil {
		return fmt.Errorf("queue: nak with delay: %w", nakErr)
	}
	return nil
}

func (q *NATSQueue) publishDLQ(ctx context.Context, m Message, attempts int) error {
	body, err := json.Marshal(map[string]any{
		"job_id":      m.JobID,
		"payload":     m.Payload,
		"attempts":    attempts,
		"dead_letter": true,
		"at":          time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return fmt.Errorf("queue: marshal dead-letter message: %w", err)
	}
	log.Printf("queue dead-letter event job_id=%s attempts=%d dead_letter=true subject=%s", m.JobID, attempts, q.cfg.DeadLetter)
	if _, err := q.js.Publish(q.cfg.DeadLetter, body, nats.Context(ctx)); err != nil {
		return fmt.Errorf("queue: publish dead-letter: %w", err)
	}
	return nil
}
