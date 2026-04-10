package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewRegistry_Handler(t *testing.T) {
	t.Parallel()
	reg := NewRegistry()
	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	s := string(body)
	for _, needle := range []string{"go_goroutines", "process_cpu_seconds_total", "go_memstats"} {
		if !strings.Contains(s, needle) {
			t.Fatalf("response missing %q", needle)
		}
	}
}

func TestWithCollector(t *testing.T) {
	t.Parallel()
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "senju_metrics_test_probe_total",
		Help: "Test counter for metrics package registration.",
	})
	reg := NewRegistry(WithCollector(counter))
	counter.Inc()

	srv := httptest.NewServer(reg.Handler())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "senju_metrics_test_probe_total") {
		t.Fatal("custom counter not exposed")
	}
}
