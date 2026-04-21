package variantquery

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/variant/clickhouse"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	gotFilters clickhouse.QueryFilters
	res        clickhouse.QueryResult
	err        error
}

func (f *fakeService) Query(_ context.Context, filters clickhouse.QueryFilters) (clickhouse.QueryResult, error) {
	f.gotFilters = filters
	if f.err != nil {
		return clickhouse.QueryResult{}, f.err
	}
	return f.res, nil
}

func TestGetVariants_200(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	svc := &fakeService{
		res: clickhouse.QueryResult{
			Rows: []clickhouse.QueryRow{
				{Chromosome: "chr1", Position: 123, Ref: "A", Alt: "G", Filter: "PASS", Info: "GENE=TP53", Gene: "TP53"},
			},
			Total:    1,
			Page:     2,
			PageSize: 10,
			HasNext:  false,
		},
	}
	r := gin.New()
	Register(r.Group("/v1"), svc)

	req := httptest.NewRequest(http.MethodGet, "/v1/variants?chromosome=chr1&gene=TP53&page=2&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	if svc.gotFilters.Chromosome != "chr1" || svc.gotFilters.Gene != "TP53" {
		t.Fatalf("filters %+v", svc.gotFilters)
	}
	var got response
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 || len(got.Data) != 1 || got.Data[0].Gene != "TP53" {
		t.Fatalf("response %+v", got)
	}
}

func TestGetVariants_400_Validation(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1"), &fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/variants?chromosome=chr1;DROP&page=0&page_size=1000&position_min=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var p problem.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Type != problem.TypeValidationError || len(p.Errors) == 0 {
		t.Fatalf("problem %+v", p)
	}
}

func TestGetVariants_400_PositionRangeInvalid(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1"), &fakeService{})

	req := httptest.NewRequest(http.MethodGet, "/v1/variants?position_min=100&position_max=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var p problem.Problem
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range p.Errors {
		if e.Field == "position_range" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected position_range error, got %+v", p.Errors)
	}
}

func TestGetVariants_500_QueryError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1"), &fakeService{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodGet, "/v1/variants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", w.Code)
	}
}

func TestGetVariants_503_NoService(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group("/v1"), nil)
	req := httptest.NewRequest(http.MethodGet, "/v1/variants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Variant query service is not available") {
		t.Fatalf("body %s", w.Body.String())
	}
}
