// Package stagemetrics provides shared Prometheus collectors for pipeline stages.
package stagemetrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics exports stage duration and total outcome/error counters.
type Metrics struct {
	duration *prometheus.HistogramVec
	total    *prometheus.CounterVec
}

// New constructs shared pipeline stage metrics collectors.
func New() *Metrics {
	return &Metrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "senju_pipeline_stage_duration_seconds",
			Help:    "Duration of pipeline stage executions by stage and outcome.",
			Buckets: prometheus.DefBuckets,
		}, []string{"stage", "outcome"}),
		total: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "senju_pipeline_stage_total",
			Help: "Total pipeline stage executions by stage, outcome, and classified error type.",
		}, []string{"stage", "outcome", "error_class"}),
	}
}

// Collectors returns collectors for registration with a metrics registry.
func (m *Metrics) Collectors() []prometheus.Collector {
	return []prometheus.Collector{m.duration, m.total}
}

// Observe records one stage execution.
func (m *Metrics) Observe(stage, outcome, errorClass string, duration time.Duration) {
	m.duration.WithLabelValues(stage, outcome).Observe(duration.Seconds())
	m.total.WithLabelValues(stage, outcome, errorClass).Inc()
}
