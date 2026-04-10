package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/testdb"
	"github.com/AminN77/senju/backend/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func TestRepository_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	base := os.Getenv("POSTGRES_DSN")
	if base == "" {
		t.Skip("POSTGRES_DSN not set")
	}

	dsn := testdb.NewIsolatedDatabase(t, base)
	applyMigrations(t, dsn)
	t.Cleanup(func() {
		if err := migrateDown(dsn); err != nil {
			t.Errorf("migrate down: %v", err)
		}
	})

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	repo := NewRepository(pool)

	created, err := repo.Create(ctx, job.CreateParams{
		Status:   job.StatusPending,
		Stage:    "qc",
		InputRef: json.RawMessage(`{"ref":"s3://bucket/in"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != job.StatusPending || created.Stage != "qc" {
		t.Fatalf("create: %+v", created)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	var gotObj, wantObj any
	if err := json.Unmarshal(got.InputRef, &gotObj); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"ref":"s3://bucket/in"}`), &wantObj); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotObj, wantObj) {
		t.Fatalf("input ref: %s", got.InputRef)
	}

	updated, err := repo.Update(ctx, created.ID, job.UpdateParams{
		Status: job.StatusRunning,
		Stage:  "align",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != job.StatusRunning || updated.Stage != "align" {
		t.Fatalf("update: %+v", updated)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("expected updated_at to advance: before=%v after=%v", created.UpdatedAt, updated.UpdatedAt)
	}
}

func applyMigrations(t *testing.T, dsn string) {
	t.Helper()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	pgDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		t.Fatal(err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", pgDriver)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatal(err)
	}
}

func migrateDown(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	pgDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", pgDriver)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Down(); err != nil {
		return err
	}
	return nil
}
