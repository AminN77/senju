package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/api/variantquery"
	"github.com/AminN77/senju/backend/internal/healthcheck"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/AminN77/senju/backend/internal/ml/impact"
	"github.com/AminN77/senju/backend/internal/platform/metrics"
	"github.com/AminN77/senju/backend/internal/security"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// testMetricsHandler returns a real Prometheus handler for Register (required non-nil Metrics).
func testMetricsHandler() http.Handler {
	return metrics.NewRegistry().Handler()
}

type testAllowAllAuth struct{}

func (testAllowAllAuth) RequireRoles(_ ...string) gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func testAuthAllowAll() security.Authorizer { return testAllowAllAuth{} }

func BenchmarkHealthLive(b *testing.B) {
	b.ReportAllocs()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	Register(r, Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "0", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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
		Log:             zerolog.Nop(),
		Auth:            testAuthAllowAll(),
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

type testVariantSvc struct{}

func (testVariantSvc) Query(_ context.Context, f variantquery.QueryFilters) (variantquery.QueryResult, error) {
	return variantquery.QueryResult{Rows: []variantquery.QueryRow{
		{Chromosome: "chr1", Position: 1, Ref: "A", Alt: "T", Filter: "PASS", Info: "GENE=TP53", Gene: "TP53"},
	}, Total: 1, Page: f.Page, PageSize: f.PageSize, HasNext: false}, nil
}

type testMLImpactSvc struct{}

func (testMLImpactSvc) Train(_ context.Context, _ []impact.TrainSample) (impact.Metadata, error) {
	return impact.Metadata{
		DatasetHash:    "h",
		FeatureVersion: impact.FeatureVersion,
		ModelVersion:   "m1",
		TrainedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		SampleCount:    2,
	}, nil
}

func (testMLImpactSvc) Predict(_ context.Context, _ impact.PredictInput) (impact.PredictResult, error) {
	return impact.PredictResult{
		Score:     0.9,
		Class:     "deleterious",
		LatencyMS: 1,
		Features: map[string]float64{
			"qual_scaled": 0.9,
		},
		Metadata: impact.Metadata{
			DatasetHash:    "h",
			FeatureVersion: impact.FeatureVersion,
			ModelVersion:   "m1",
			TrainedAt:      time.Now().UTC().Format(time.RFC3339Nano),
			SampleCount:    2,
		},
	}, nil
}

func TestAuthProtection_VariantsRequiresAnalystRole(t *testing.T) {
	t.Parallel()
	authz, err := security.NewJWTAuthorizer("secret", "senju-api")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
		Log:             zerolog.Nop(),
		Auth:            authz,
		VariantQuery:    testVariantSvc{},
		MLImpact:        testMLImpactSvc{},
	}))
	t.Cleanup(srv.Close)

	t.Run("missing token", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/v1/variants")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/v1/variants", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testSignedJWT(t, "secret", "senju-api", "uploader"))
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})

	t.Run("allowed role", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/v1/variants", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testSignedJWT(t, "secret", "senju-api", "analyst"))
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})
}

func TestAuthProtection_MLImpactRequiresAnalystRole(t *testing.T) {
	t.Parallel()
	authz, err := security.NewJWTAuthorizer("secret", "senju-api")
	if err != nil {
		t.Fatal(err)
	}
	repo := mlimpactTestRepo(t)
	srv := httptest.NewServer(testEngine(Options{
		Readiness:       healthcheck.NewRunner(),
		Version:         VersionInfo{Service: "senju-api", Version: "test", Commit: "x", BuildTime: "y"},
		EnableSwaggerUI: false,
		Metrics:         testMetricsHandler(),
		Log:             zerolog.Nop(),
		Auth:            authz,
		MLImpact:        testMLImpactSvc{},
		Jobs:            repo,
	}))
	t.Cleanup(srv.Close)
	jobID := createMLTestJob(t, repo)
	body := `{"variant":{"chromosome":"chr1","position":1,"ref":"A","alt":"T","qual":90,"filter":"PASS","gene":"TP53"}}`
	t.Run("missing token", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/ml/impact/"+jobID+"/predict", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})
	t.Run("wrong role", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/ml/impact/"+jobID+"/predict", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testSignedJWT(t, "secret", "senju-api", "runner"))
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})
	t.Run("allowed role", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/ml/impact/"+jobID+"/predict", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testSignedJWT(t, "secret", "senju-api", "analyst"))
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status=%d", resp.StatusCode)
		}
	})
}

func testSignedJWT(t *testing.T, secret, issuer, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  uuid.NewString(),
		"iss":  issuer,
		"exp":  time.Now().Add(30 * time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"role": role,
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func mlimpactTestRepo(t *testing.T) job.Repository {
	t.Helper()
	return stub.New()
}

func createMLTestJob(t *testing.T, repo job.Repository) string {
	t.Helper()
	j, err := repo.Create(context.Background(), job.CreateParams{
		Status: job.StatusRunning,
		Stage:  "pipeline_running",
	})
	if err != nil {
		t.Fatal(err)
	}
	return j.ID.String()
}
