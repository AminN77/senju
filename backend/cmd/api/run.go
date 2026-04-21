package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AminN77/senju/backend/internal/config"
	"github.com/AminN77/senju/backend/internal/healthcheck"
	"github.com/AminN77/senju/backend/internal/httpserver"
	"github.com/AminN77/senju/backend/internal/job"
	jobpostgres "github.com/AminN77/senju/backend/internal/job/postgres"
	"github.com/AminN77/senju/backend/internal/objectstore"
	"github.com/AminN77/senju/backend/internal/platform/logging"
	"github.com/AminN77/senju/backend/internal/platform/metrics"
	"github.com/AminN77/senju/backend/internal/probe/httpget"
	"github.com/AminN77/senju/backend/internal/variant/clickhouse"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 1 << 20 // 1 MiB
)

// run loads configuration, builds the HTTP handler, serves until SIGINT/SIGTERM, then shuts down gracefully.
func run() error {
	setGinModeFromEnv()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log := logging.New(cfg.LogLevel)
	log.Info().Str("gin_mode", gin.Mode()).Msg("gin mode")

	runner := healthcheck.NewRunner()
	if err := registerReadinessProbes(runner, cfg, httpget.DefaultClient()); err != nil {
		return err
	}
	promRegistry := metrics.NewRegistry()

	var jobRepo job.Repository
	var pgPool *pgxpool.Pool
	if cfg.PostgresDSN != "" {
		pool, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
		if err != nil {
			return fmt.Errorf("postgres pool: %w", err)
		}
		pgPool = pool
		jobRepo = jobpostgres.NewRepository(pool)
	}
	defer func() {
		if pgPool != nil {
			pgPool.Close()
		}
	}()

	var objStore objectstore.Service
	if cfg.ObjectStore.Enabled() {
		s, err := objectstore.NewS3(objectstore.S3Options{
			Endpoint:     cfg.ObjectStore.Endpoint,
			Region:       cfg.ObjectStore.Region,
			Bucket:       cfg.ObjectStore.Bucket,
			AccessKey:    cfg.ObjectStore.AccessKey,
			SecretKey:    cfg.ObjectStore.SecretKey,
			UsePathStyle: cfg.ObjectStore.UsePathStyle,
		})
		if err != nil {
			return fmt.Errorf("object store: %w", err)
		}
		objStore = s
	}

	var variantQuery clickhouse.QueryService
	var variantRepo *clickhouse.QueryRepository
	if cfg.ClickHouseDSN != "" {
		repo, err := clickhouse.OpenQueryRepository(cfg.ClickHouseDSN)
		if err != nil {
			return fmt.Errorf("clickhouse query repository: %w", err)
		}
		variantRepo = repo
		variantQuery = repo
	}
	defer func() {
		if variantRepo != nil {
			_ = variantRepo.Close()
		}
	}()

	engine := newEngine(log, runner, versionInfo(), promRegistry, jobRepo, objStore, variantQuery)
	addr := listenAddr(cfg.APIPort)

	srv := &http.Server{
		Addr:              addr,
		Handler:           engine,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	log.Info().Str("addr", addr).Msg("senju-api listening")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutdown requested")
	}

	shutdownTimeout := parseShutdownTimeout(log)
	shCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	log.Info().Msg("server stopped")
	return nil
}

// setGinModeFromEnv is the single source of truth after Gin's init (which also reads GIN_MODE).
// We normalize with TrimSpace + lowercase so " debug " and DEBUG behave predictably.
// Empty, release, production → release. debug → debug. test → test. Unknown → release (and Gin would
// have panicked at init on an invalid value like "foo" — only use debug, release, or test).
func setGinModeFromEnv() {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GIN_MODE"))) {
	case "", "release", "production":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

func listenAddr(port int) string {
	return ":" + strconv.Itoa(port)
}

func newEngine(log zerolog.Logger, readiness httpserver.ReadinessChecker, ver httpserver.VersionInfo, prom *metrics.Registry, jobs job.Repository, store objectstore.Service, variants clickhouse.QueryService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(httpserver.RequestLogger(log))
	httpserver.Register(r, httpserver.Options{
		Readiness:       readiness,
		Version:         ver,
		EnableSwaggerUI: gin.Mode() != gin.ReleaseMode,
		Metrics:         prom.Handler(),
		Jobs:            jobs,
		Log:             log,
		ObjectStore:     store,
		VariantQuery:    variants,
	})
	return r
}

func parseShutdownTimeout(log zerolog.Logger) time.Duration {
	const defaultTimeout = 30 * time.Second
	const minTimeout = time.Second

	s := os.Getenv("SHUTDOWN_TIMEOUT")
	if s == "" {
		return defaultTimeout
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Warn().Str("SHUTDOWN_TIMEOUT", s).Err(err).Msg("invalid duration; using default")
		return defaultTimeout
	}
	if d < minTimeout {
		return minTimeout
	}
	return d
}
