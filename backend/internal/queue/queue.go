// Package queue defines transport-agnostic job queue contracts.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Message is the queue payload for one workflow job dispatch.
type Message struct {
	JobID   string          `json:"job_id"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Handler processes one dequeued message. Returning nil acknowledges success.
// Returning a non-nil error schedules retry/dead-letter behavior.
type Handler func(ctx context.Context, m Message) error

// Queue supports enqueue and long-running consumer loop.
type Queue interface {
	Enqueue(ctx context.Context, m Message) error
	Consume(ctx context.Context, h Handler) error
}

// Config controls retry and subject naming behavior.
type Config struct {
	StreamName   string
	Subject      string
	ConsumerName string
	DeadLetter   string
	MaxRetries   int
	BackoffBase  time.Duration
}

// Validate checks required queue configuration.
func (c Config) Validate() error {
	if c.StreamName == "" {
		return errors.New("queue config: stream_name is required")
	}
	if c.Subject == "" {
		return errors.New("queue config: subject is required")
	}
	if c.ConsumerName == "" {
		return errors.New("queue config: consumer_name is required")
	}
	if c.DeadLetter == "" {
		return errors.New("queue config: dead_letter is required")
	}
	if c.MaxRetries < 0 {
		return errors.New("queue config: max_retries must be >= 0")
	}
	if c.BackoffBase <= 0 {
		return errors.New("queue config: backoff_base must be > 0")
	}
	return nil
}

// RetryDelay returns the backoff delay for the given retry number (1-based).
func (c Config) RetryDelay(retryNum int) time.Duration {
	if retryNum <= 1 {
		return c.BackoffBase
	}
	return c.BackoffBase * time.Duration(1<<(retryNum-1))
}
