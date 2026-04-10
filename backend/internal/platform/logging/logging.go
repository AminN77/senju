// Package logging configures structured logging with zerolog for the API process.
//
// Zerolog is chosen for zero-allocation JSON logging and a stable, minimal API;
// see docs/backend-engineering.md (library selection) and module pins in go.mod.
package logging

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// New returns a process logger writing to stdout. Level uses zerolog.ParseLevel; empty defaults to info.
func New(level string) zerolog.Logger {
	lvl := zerolog.InfoLevel
	if s := strings.TrimSpace(level); s != "" {
		if parsed, err := zerolog.ParseLevel(s); err == nil {
			lvl = parsed
		}
	}
	zerolog.SetGlobalLevel(lvl)
	out := io.Writer(os.Stdout)
	return zerolog.New(out).With().Timestamp().Logger()
}
