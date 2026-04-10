// Package postgres provides a generic SQL ping [healthcheck.Probe] for PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AminN77/senju/backend/internal/healthcheck"

	// Driver for database/sql.
	_ "github.com/lib/pq"
)

// Probe pings PostgreSQL using a DSN. Name identifies this instance in readiness output.
type Probe struct {
	name string
	dsn  string
}

// New returns a [healthcheck.Probe] that verifies connectivity with PingContext.
func New(name, dsn string) healthcheck.Probe {
	return &Probe{name: name, dsn: dsn}
}

// Name implements [healthcheck.Probe].
func (p *Probe) Name() string { return p.name }

// Check implements [healthcheck.Probe].
func (p *Probe) Check(ctx context.Context) error {
	if p.dsn == "" {
		return fmt.Errorf("dsn not configured")
	}
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer func() { _ = db.Close() }()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}
