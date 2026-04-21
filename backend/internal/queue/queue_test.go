package queue

import (
	"testing"
	"time"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()
	cfg := Config{
		StreamName:   "jobs",
		Subject:      "jobs.execute",
		ConsumerName: "worker",
		DeadLetter:   "jobs.dead_letter",
		MaxRetries:   3,
		BackoffBase:  50 * time.Millisecond,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func TestRetryDelay(t *testing.T) {
	t.Parallel()
	cfg := Config{BackoffBase: 100 * time.Millisecond}
	if got := cfg.RetryDelay(1); got != 100*time.Millisecond {
		t.Fatalf("attempt1 got=%s", got)
	}
	if got := cfg.RetryDelay(2); got != 200*time.Millisecond {
		t.Fatalf("attempt2 got=%s", got)
	}
	if got := cfg.RetryDelay(3); got != 400*time.Millisecond {
		t.Fatalf("attempt3 got=%s", got)
	}
}
