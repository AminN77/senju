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
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.ClickHousePing != "" {
			t.Fatalf("got %q", cfg.ClickHousePing)
		}
	})

	t.Run("with base", func(t *testing.T) {
		t.Setenv("CLICKHOUSE_HTTP_URL", "http://localhost:8123")
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if want := "http://localhost:8123/ping"; cfg.ClickHousePing != want {
			t.Fatalf("got %q want %q", cfg.ClickHousePing, want)
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
