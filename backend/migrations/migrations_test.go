package migrations

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/AminN77/senju/backend/internal/testdb"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

func TestEmbed_SQLFilesPresent(t *testing.T) {
	t.Parallel()
	names := []string{
		"000001_jobs.up.sql",
		"000001_jobs.down.sql",
	}
	for _, name := range names {
		b, err := fs.ReadFile(Files, name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if len(b) == 0 {
			t.Fatalf("empty %s", name)
		}
	}
}

func TestMigrateUpDown(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	base := os.Getenv("POSTGRES_DSN")
	if base == "" {
		t.Skip("POSTGRES_DSN not set")
	}

	dsn := testdb.NewIsolatedDatabase(t, base)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	pgDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	src, err := iofs.New(Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", pgDriver)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("up: %v", err)
	}
	v, dirty, err := m.Version()
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Fatalf("version dirty after up: %d", v)
	}
	if v != 1 {
		t.Fatalf("version: got %d want 1", v)
	}

	var exists bool
	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'jobs'
		)`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("jobs table missing after up")
	}

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("down: %v", err)
	}

	if _, _, err := m.Version(); err == nil {
		t.Fatal("expected error from Version() after full down")
	} else if !errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("Version after down: %v", err)
	}

	if err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'jobs'
		)`).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("jobs table still exists after down")
	}
}
