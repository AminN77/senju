// Package testdb provides helpers for Postgres-backed integration tests.
package testdb

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/lib/pq"

	_ "github.com/lib/pq" // register driver for database/sql
)

// NewIsolatedDatabase creates a new empty database and returns a DSN to it.
// The database is dropped in t.Cleanup (PostgreSQL 13+ DROP ... WITH (FORCE)).
//
// Requires a role that can CREATE DATABASE (e.g. CI superuser). For local
// Compose, use a superuser DSN or grant CREATEDB to the app role if tests fail.
func NewIsolatedDatabase(t *testing.T, baseDSN string) string {
	t.Helper()
	u, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	dbName := fmt.Sprintf("senju_test_%d", time.Now().UnixNano())
	admin := *u
	admin.Path = "/postgres"

	adminDB, err := sql.Open("postgres", admin.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = adminDB.Close() }()

	if _, err := adminDB.Exec("CREATE DATABASE " + pq.QuoteIdentifier(dbName)); err != nil {
		t.Fatalf("create database: %v", err)
	}

	out := *u
	out.Path = "/" + dbName
	dsn := out.String()

	t.Cleanup(func() {
		adm, err := sql.Open("postgres", admin.String())
		if err != nil {
			t.Errorf("testdb cleanup open: %v", err)
			return
		}
		defer func() { _ = adm.Close() }()
		_, err = adm.Exec(`DROP DATABASE IF EXISTS ` + pq.QuoteIdentifier(dbName) + ` WITH (FORCE)`)
		if err != nil {
			t.Errorf("drop test database %s: %v", dbName, err)
		}
	})
	return dsn
}
