// Package postgres provides a generic SQL ping [healthcheck.Probe] for PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/healthcheck"

	// Driver for database/sql.
	_ "github.com/lib/pq"
)

// Probe pings PostgreSQL using a long-lived pool opened once. Name identifies this instance in readiness output.
type Probe struct {
	name string
	db   *sql.DB
}

// New opens a single connection pool for the DSN and returns a [healthcheck.Probe] that uses PingContext.
// The pool is retained for the process lifetime; close is not required for short-lived API processes.
func New(name, dsn string) (healthcheck.Probe, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, fmt.Errorf("dsn not configured")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	return &Probe{name: name, db: db}, nil
}

// Name implements [healthcheck.Probe].
func (p *Probe) Name() string { return p.name }

// Check implements [healthcheck.Probe].
func (p *Probe) Check(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := p.db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}
