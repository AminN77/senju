package config

import (
	"strings"
	"testing"
)

func TestLoad_APIPort(t *testing.T) {
	tests := []struct {
		name       string
		apiPort    string
		wantPort   int
		wantErr    bool
		errContain string
	}{
		{
			name:     "default when unset",
			apiPort:  "",
			wantPort: 8080,
		},
		{
			name:     "valid custom port",
			apiPort:  "9090",
			wantPort: 9090,
		},
		{
			name:       "invalid not a number",
			apiPort:    "not-a-port",
			wantErr:    true,
			errContain: "API_PORT",
		},
		{
			name:       "out of range high",
			apiPort:    "65536",
			wantErr:    true,
			errContain: "API_PORT",
		},
		{
			name:       "out of range zero",
			apiPort:    "0",
			wantErr:    true,
			errContain: "API_PORT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("API_PORT", tt.apiPort)
			t.Setenv("CLICKHOUSE_HTTP_URL", "")
			t.Setenv("POSTGRES_HOST", "")
			t.Setenv("POSTGRES_DSN", "")

			got, err := Load()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Fatalf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.APIPort != tt.wantPort {
				t.Fatalf("APIPort: got %d want %d", got.APIPort, tt.wantPort)
			}
		})
	}
}

func TestLoad_ClickHousePingOnlyWhenBaseSet(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_DSN", "")

	t.Run("empty base", func(t *testing.T) {
		t.Setenv("CLICKHOUSE_HTTP_URL", "")
		t.Setenv("CLICKHOUSE_DSN", "")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ClickHousePing != "" {
			t.Fatalf("got %q", cfg.ClickHousePing)
		}
		if cfg.ClickHouseDSN != "" {
			t.Fatalf("got %q", cfg.ClickHouseDSN)
		}
	})

	t.Run("with base", func(t *testing.T) {
		t.Setenv("CLICKHOUSE_HTTP_URL", "http://localhost:8123")
		t.Setenv("CLICKHOUSE_NATIVE_PORT", "9000")
		t.Setenv("CLICKHOUSE_USER", "default")
		t.Setenv("CLICKHOUSE_PASSWORD", "pass")
		t.Setenv("CLICKHOUSE_DB", "senju")
		t.Setenv("CLICKHOUSE_DSN", "")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if want := "http://localhost:8123/ping"; cfg.ClickHousePing != want {
			t.Fatalf("got %q want %q", cfg.ClickHousePing, want)
		}
		if !strings.Contains(cfg.ClickHouseDSN, "clickhouse://default:pass@localhost:9000/senju") {
			t.Fatalf("dsn %q", cfg.ClickHouseDSN)
		}
	})

	t.Run("explicit dsn overrides base", func(t *testing.T) {
		t.Setenv("CLICKHOUSE_HTTP_URL", "http://localhost:8123")
		t.Setenv("CLICKHOUSE_DSN", "clickhouse://u:p@localhost:9000/db")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ClickHouseDSN != "clickhouse://u:p@localhost:9000/db" {
			t.Fatalf("dsn %q", cfg.ClickHouseDSN)
		}
	})
}

func TestLoad_ObjectStoreEnabled(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("CLICKHOUSE_HTTP_URL", "")
	t.Setenv("S3_REGION", "")
	t.Setenv("S3_ENDPOINT", "http://127.0.0.1:9000")
	t.Setenv("S3_BUCKET", "test-bucket")
	t.Setenv("S3_ACCESS_KEY", "access")
	t.Setenv("S3_SECRET_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.ObjectStore.Enabled() {
		t.Fatal("expected object store enabled")
	}
	if cfg.ObjectStore.Region != "us-east-1" {
		t.Fatalf("region: got %q", cfg.ObjectStore.Region)
	}
}

func TestLoad_QueueDefaultsAndOverrides(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	t.Setenv("POSTGRES_HOST", "")
	t.Setenv("POSTGRES_DSN", "")
	t.Setenv("CLICKHOUSE_HTTP_URL", "")

	t.Run("defaults", func(t *testing.T) {
		t.Setenv("QUEUE_STREAM_NAME", "")
		t.Setenv("QUEUE_SUBJECT", "")
		t.Setenv("QUEUE_DEAD_LETTER_SUBJECT", "")
		t.Setenv("QUEUE_CONSUMER_NAME", "")
		t.Setenv("QUEUE_MAX_RETRIES", "")
		t.Setenv("QUEUE_BACKOFF_BASE", "")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Queue.StreamName != "jobs_stream" || cfg.Queue.Subject != "jobs.execute" {
			t.Fatalf("defaults queue %+v", cfg.Queue)
		}
		if cfg.Queue.DeadLetter != "jobs.dead_letter" || cfg.Queue.ConsumerName != "jobs_worker" {
			t.Fatalf("defaults queue %+v", cfg.Queue)
		}
		if cfg.Queue.MaxRetries != 3 {
			t.Fatalf("max retries got %d", cfg.Queue.MaxRetries)
		}
	})

	t.Run("overrides", func(t *testing.T) {
		t.Setenv("QUEUE_STREAM_NAME", "s")
		t.Setenv("QUEUE_SUBJECT", "work")
		t.Setenv("QUEUE_DEAD_LETTER_SUBJECT", "dlq")
		t.Setenv("QUEUE_CONSUMER_NAME", "worker")
		t.Setenv("QUEUE_MAX_RETRIES", "8")
		t.Setenv("QUEUE_BACKOFF_BASE", "250ms")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Queue.StreamName != "s" || cfg.Queue.Subject != "work" {
			t.Fatalf("queue %+v", cfg.Queue)
		}
		if cfg.Queue.DeadLetter != "dlq" || cfg.Queue.ConsumerName != "worker" {
			t.Fatalf("queue %+v", cfg.Queue)
		}
		if cfg.Queue.MaxRetries != 8 {
			t.Fatalf("max retries got %d", cfg.Queue.MaxRetries)
		}
		if got := cfg.Queue.BackoffBase.String(); got != "250ms" {
			t.Fatalf("backoff got %s", got)
		}
	})

	t.Run("malformed numeric and duration env", func(t *testing.T) {
		t.Setenv("QUEUE_STREAM_NAME", "s")
		t.Setenv("QUEUE_SUBJECT", "work")
		t.Setenv("QUEUE_DEAD_LETTER_SUBJECT", "dlq")
		t.Setenv("QUEUE_CONSUMER_NAME", "worker")
		t.Setenv("QUEUE_MAX_RETRIES", "abc")
		t.Setenv("QUEUE_BACKOFF_BASE", "250ms")

		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "QUEUE_MAX_RETRIES") {
			t.Fatalf("expected QUEUE_MAX_RETRIES parse error, got %v", err)
		}

		t.Setenv("QUEUE_MAX_RETRIES", "3")
		t.Setenv("QUEUE_BACKOFF_BASE", "oops")
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "QUEUE_BACKOFF_BASE") {
			t.Fatalf("expected QUEUE_BACKOFF_BASE parse error, got %v", err)
		}
	})

	t.Run("whitespace-only queue fields use defaults", func(t *testing.T) {
		t.Setenv("QUEUE_STREAM_NAME", "   ")
		t.Setenv("QUEUE_SUBJECT", " ")
		t.Setenv("QUEUE_DEAD_LETTER_SUBJECT", "\t")
		t.Setenv("QUEUE_CONSUMER_NAME", "   ")
		t.Setenv("QUEUE_MAX_RETRIES", "")
		t.Setenv("QUEUE_BACKOFF_BASE", "")

		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Queue.StreamName != "jobs_stream" || cfg.Queue.Subject != "jobs.execute" {
			t.Fatalf("queue defaults %+v", cfg.Queue)
		}
		if cfg.Queue.DeadLetter != "jobs.dead_letter" || cfg.Queue.ConsumerName != "jobs_worker" {
			t.Fatalf("queue defaults %+v", cfg.Queue)
		}
	})
}
