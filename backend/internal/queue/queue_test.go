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

func TestConfigValidate_InvalidCases(t *testing.T) {
	t.Parallel()

	valid := Config{
		StreamName:   "jobs",
		Subject:      "jobs.execute",
		ConsumerName: "worker",
		DeadLetter:   "jobs.dead_letter",
		MaxRetries:   3,
		BackoffBase:  time.Second,
	}

	tests := []struct {
		name string
		cfg  Config
	}{
		{name: "missing stream name", cfg: func() Config { c := valid; c.StreamName = ""; return c }()},
		{name: "whitespace stream name", cfg: func() Config { c := valid; c.StreamName = "   "; return c }()},
		{name: "missing subject", cfg: func() Config { c := valid; c.Subject = ""; return c }()},
		{name: "whitespace subject", cfg: func() Config { c := valid; c.Subject = "\t"; return c }()},
		{name: "missing consumer", cfg: func() Config { c := valid; c.ConsumerName = ""; return c }()},
		{name: "whitespace consumer", cfg: func() Config { c := valid; c.ConsumerName = " "; return c }()},
		{name: "missing dead-letter", cfg: func() Config { c := valid; c.DeadLetter = ""; return c }()},
		{name: "whitespace dead-letter", cfg: func() Config { c := valid; c.DeadLetter = "  "; return c }()},
		{name: "negative max retries", cfg: func() Config { c := valid; c.MaxRetries = -1; return c }()},
		{name: "zero backoff base", cfg: func() Config { c := valid; c.BackoffBase = 0; return c }()},
		{name: "negative backoff base", cfg: func() Config { c := valid; c.BackoffBase = -time.Second; return c }()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.cfg.Validate(); err == nil {
				t.Fatal("expected validate error")
			}
		})
	}
}

func TestRetryDelay_Edges(t *testing.T) {
	t.Parallel()
	cfg := Config{BackoffBase: 10 * time.Millisecond}
	if got := cfg.RetryDelay(0); got != 10*time.Millisecond {
		t.Fatalf("attempt0 got=%s", got)
	}
	if got := cfg.RetryDelay(-5); got != 10*time.Millisecond {
		t.Fatalf("attempt-5 got=%s", got)
	}
	if got := cfg.RetryDelay(8); got != 1280*time.Millisecond {
		t.Fatalf("attempt8 got=%s", got)
	}
}
