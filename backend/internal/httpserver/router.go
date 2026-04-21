// Package httpserver registers Gin routes for the Senju HTTP API control plane.
//
// Readiness is expressed as [ReadinessChecker] so this package depends on a boundary
// interface rather than a concrete probe runner (see docs/backend-engineering.md).
package httpserver

import (
	"context"
	"net/http"
	"reflect"
	"time"

	"github.com/AminN77/senju/backend/internal/api/fastqupload"
	"github.com/AminN77/senju/backend/internal/api/objectupload"
	"github.com/AminN77/senju/backend/internal/api/variantquery"
	"github.com/AminN77/senju/backend/internal/healthcheck"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/objectstore"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ReadinessChecker aggregates readiness probes. Typically *healthcheck.Runner.
// Do not assign a nil *healthcheck.Runner to a ReadinessChecker field (Go interface typed-nil gotcha).
type ReadinessChecker interface {
	Check(ctx context.Context) healthcheck.Result
}

// Options configures routes and dependencies.
type Options struct {
	Readiness ReadinessChecker
	Version   VersionInfo
	// EnableSwaggerUI registers GET /docs (Swagger UI). Should be false in Gin release mode (e.g. production).
	EnableSwaggerUI bool
	// Metrics exposes Prometheus text exposition at GET /metrics (required; use platform/metrics Registry.Handler() in production).
	Metrics http.Handler
	// Jobs persists job metadata. If nil, FASTQ metadata routes return 503 (database not configured).
	Jobs job.Repository
	// Log is used for structured request logging and object upload audit lines.
	Log zerolog.Logger
	// ObjectStore enables multipart presigned uploads under /v1/objects. If nil, those routes return 503.
	ObjectStore objectstore.Service
	// VariantQuery serves variant analytics reads. If nil, /v1/variants returns 503.
	VariantQuery variantquery.Service
}

// VersionInfo is returned by GET /version.
type VersionInfo struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// Register mounts API routes on the Gin engine (see ADR-0004 and OpenAPI spec).
// opts.Readiness and opts.Metrics must be non-nil (typed-nil interfaces are rejected for Readiness).
func Register(r *gin.Engine, opts Options) {
	if isNilReadinessChecker(opts.Readiness) {
		panic("httpserver: Options.Readiness must not be nil")
	}
	if opts.Metrics == nil {
		panic("httpserver: Options.Metrics must not be nil")
	}
	r.GET("/health/live", handleLive)
	r.GET("/health/ready", handleReady(opts.Readiness))
	r.GET("/version", handleVersion(opts.Version))
	r.GET("/", handleRoot)
	r.GET("/metrics", gin.WrapH(opts.Metrics))
	fastqupload.Register(r.Group("/v1/jobs"), opts.Jobs)
	objectupload.Register(r.Group("/v1/objects"), opts.ObjectStore, opts.Log)
	variantquery.Register(r.Group("/v1"), opts.VariantQuery)
	registerOpenAPISpecRoute(r)
	if opts.EnableSwaggerUI {
		registerSwaggerUIRoute(r)
	}
}

// RequestLogger emits one structured log line per request (method, path, status, latency).
func RequestLogger(log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Msg("http")
	}
}

func handleLive(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("ok\n"))
}

func handleReady(checker ReadinessChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		res := checker.Check(c.Request.Context())
		body := healthcheck.FormatBody(res)
		if res.OK {
			c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
			return
		}
		c.Data(http.StatusServiceUnavailable, "text/plain; charset=utf-8", []byte(body))
	}
}

func handleVersion(v VersionInfo) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, v)
	}
}

func handleRoot(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("senju api\n"))
}

func isNilReadinessChecker(v ReadinessChecker) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}
