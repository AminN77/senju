package httpserver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/healthcheck"
	"github.com/AminN77/senju/backend/internal/platform/metrics"
	"github.com/gin-gonic/gin"
)

// testMetricsHandler returns a real Prometheus handler for Register (required non-nil Metrics).
func testMetricsHandler() http.Handler {
	return metrics.NewRegistry().Handler()
}

func BenchmarkHealthLive(b *testing.B) {
	b.ReportAllocs()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	Register(r, Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "0", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
	})
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("status %d", w.Code)
		}
	}
}

func BenchmarkHealthReady_EmptyRunner(b *testing.B) {
	b.ReportAllocs()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	Register(r, Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "0", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
	})
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("status %d", w.Code)
		}
	}
}

func testEngine(opts Options) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, opts)
	return r
}

func TestHealthLive(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: true,
		Metrics:         testMetricsHandler(),
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/health/live")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
}

func TestVersionJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(testEngine(Options{
		Readiness: healthcheck.NewRunner(),
		Version: VersionInfo{
			Service: "senju-api", Version: "1.2.3", Commit: "abc", BuildTime: "now",
		},
		EnableSwaggerUI: true,
		Metrics:         testMetricsHandler(),
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/version")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	var got VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Version != "1.2.3" || got.Service != "senju-api" {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestOpenAPISpecYAML(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "yaml") {
		t.Fatalf("Content-Type: got %q", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("openapi:")) {
		t.Fatalf("body missing openapi key")
	}
}

func TestMetrics(t *testing.T) {
	t.Parallel()
	reg := metrics.NewRegistry()
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         reg.Handler(),
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("go_goroutines")) {
		t.Fatalf("metrics body missing go_goroutines")
	}
}

func TestSwaggerUI(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: true,
		Metrics:         testMetricsHandler(),
	}))
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/docs")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("swagger-ui")) {
		t.Fatalf("page should reference swagger-ui assets")
	}
}

func TestSwaggerUI_NotRegisteredWhenDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	Register(r, Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /docs when disabled, got %d", w.Code)
	}
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for /openapi.yaml when Swagger UI disabled, got %d", w2.Code)
	}
}

func TestHealthLiveP95Under100ms(t *testing.T) {
	if testing.Short() {
		t.Skip("timing threshold is environment-dependent; skipped with -short")
	}
	if os.Getenv("CI") != "" {
		t.Skip("timing threshold is environment-dependent; run locally or use benchmarks")
	}
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: true,
		Metrics:         testMetricsHandler(),
	}))
	t.Cleanup(srv.Close)

	const warmup = 50
	const samples = 500
	client := srv.Client()
	u := srv.URL + "/health/live"

	for range warmup {
		resp, err := client.Get(u)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadAll(resp.Body); err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}

	lat := make([]time.Duration, 0, samples)
	for range samples {
		start := time.Now()
		resp, err := client.Get(u)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadAll(resp.Body); err != nil {
			t.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			t.Fatal(err)
		}
		lat = append(lat, time.Since(start))
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	p95 := lat[int(float64(len(lat))*0.95)]
	if p95 > 100*time.Millisecond {
		t.Fatalf("p95 latency %s exceeds 100ms", p95)
	}
}
