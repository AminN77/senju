// Package migrations holds embedded SQL migration files for golang-migrate and tooling.
package migrations

import "embed"

// Files contains all *.sql migration pairs (up/down).
//
//go:embed *.sql
var Files embed.FS
