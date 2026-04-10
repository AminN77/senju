// Package db holds sqlc query definitions. Generated code lives in [jobdb].
//
// Regenerate after editing SQL (run from repo anywhere; paths are fixed):
//
//	go generate github.com/AminN77/senju/backend/internal/job/db
package db

//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate -f ../../../sqlc.yaml
