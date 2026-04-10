package main

import (
	"fmt"
	"net/http"

	"github.com/AminN77/senju/backend/internal/config"
	"github.com/AminN77/senju/backend/internal/healthcheck"
	"github.com/AminN77/senju/backend/internal/probe/httpget"
	"github.com/AminN77/senju/backend/internal/probe/postgres"
	"github.com/AminN77/senju/backend/internal/probe/tcp"
)

// registerReadinessProbes attaches probes derived from config. Only non-empty settings produce checks.
func registerReadinessProbes(r *healthcheck.Runner, cfg config.Config, hc *http.Client) error {
	if cfg.PostgresDSN != "" {
		p, err := postgres.New("postgres", cfg.PostgresDSN)
		if err != nil {
			return fmt.Errorf("postgres probe: %w", err)
		}
		r.Register(p)
	}
	if cfg.ClickHousePing != "" {
		r.Register(httpget.New("clickhouse", cfg.ClickHousePing, hc))
	}
	if cfg.MinIOHealthURL != "" {
		r.Register(httpget.New("minio", cfg.MinIOHealthURL, hc))
	}
	if cfg.NATSAddr != "" {
		r.Register(tcp.New("nats", "tcp", cfg.NATSAddr))
	}
	return nil
}
