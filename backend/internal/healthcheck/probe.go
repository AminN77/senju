// Package healthcheck defines readiness probes and aggregates their results.
package healthcheck

import (
	"context"
	"strings"
)

// Probe checks one dependency. Implementations are registered on a [Runner] at composition time.
type Probe interface {
	// Name identifies the probe in logs and response bodies (e.g. "postgres", "queue").
	Name() string
	// Check returns nil if the dependency is ready.
	Check(ctx context.Context) error
}

// Result is the outcome of running all probes on a [Runner].
type Result struct {
	OK     bool
	Errors []string
}

// FormatBody returns a plain-text body for HTTP readiness handlers.
func FormatBody(r Result) string {
	if r.OK {
		return "ok\n"
	}
	return "not ready:\n- " + strings.Join(r.Errors, "\n- ") + "\n"
}
