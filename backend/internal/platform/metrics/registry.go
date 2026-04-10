// Package metrics provides Prometheus instrumentation for the Senju API process.
//
// Use [NewRegistry] to obtain an isolated registry with standard runtime and process
// collectors; register application-specific metrics via [Option] or [Registry.MustRegister].
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry wraps a prometheus.Registry with a fixed set of default collectors and a
// stable HTTP exposition handler. It is safe for concurrent use by Prometheus scrapers.
type Registry struct {
	reg *prometheus.Registry
}

// Option configures [NewRegistry] after default collectors are registered.
type Option func(*Registry)

// NewRegistry returns a dedicated Prometheus registry with:
//   - Go runtime metrics (goroutines, GC, memory classes, etc.)
//   - process metrics (CPU, RSS, fds where supported)
//   - build_info (module version, Go version, checksum)
//
// Additional collectors can be registered via [Option], [Registry.MustRegister], or [Registry.Register].
func NewRegistry(opts ...Option) *Registry {
	r := &Registry{reg: prometheus.NewRegistry()}
	r.registerDefaults()
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *Registry) registerDefaults() {
	r.reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)
}

// WithCollector registers an extra collector after defaults (e.g. business or subsystem metrics).
func WithCollector(c prometheus.Collector) Option {
	return func(r *Registry) {
		r.reg.MustRegister(c)
	}
}

// Register adds c to the registry. An error is returned if registration fails (e.g. duplicate).
func (r *Registry) Register(c prometheus.Collector) error {
	return r.reg.Register(c)
}

// MustRegister registers collectors and panics on failure.
func (r *Registry) MustRegister(c ...prometheus.Collector) {
	r.reg.MustRegister(c...)
}

// Prometheus returns the underlying registry for advanced composition (e.g. custom gatherers).
func (r *Registry) Prometheus() *prometheus.Registry {
	return r.reg
}

// Handler serves Prometheus text exposition (and OpenMetrics when negotiated).
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
