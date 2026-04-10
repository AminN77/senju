// Command migrate applies golang-migrate SQL revisions using the same DSN rules as the API ([config.Load]).
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"github.com/AminN77/senju/backend/internal/config"
	"github.com/AminN77/senju/backend/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	dsn := cfg.PostgresDSN
	if dsn == "" {
		return fmt.Errorf("postgres DSN empty: set POSTGRES_DSN or POSTGRES_HOST with credentials")
	}

	args := os.Args[1:]
	if len(args) < 1 {
		return fmt.Errorf("usage: migrate <up|down|version|force> [n|version]")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("sql open: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	pgDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate postgres driver: %w", err)
	}

	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("migrate source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", pgDriver)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	switch args[0] {
	case "up":
		n := -1
		if len(args) > 1 {
			n, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("up: parse n: %w", err)
			}
		}
		if n < 0 {
			err = m.Up()
		} else {
			err = m.Steps(n)
		}
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no change")
			return nil
		}
		fmt.Println("up ok")
		return nil

	case "down":
		n := -1
		if len(args) > 1 {
			n, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("down: parse n: %w", err)
			}
		}
		if n < 0 {
			err = m.Down()
		} else {
			err = m.Steps(-n)
		}
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no change")
			return nil
		}
		fmt.Println("down ok")
		return nil

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("version: nil (no migrations applied)")
				return nil
			}
			return err
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil

	case "force":
		if len(args) < 2 {
			return fmt.Errorf("force: need version")
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("force: parse version: %w", err)
		}
		if err := m.Force(v); err != nil {
			return err
		}
		fmt.Println("force ok")
		return nil

	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
