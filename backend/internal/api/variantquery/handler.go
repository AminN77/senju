// Package variantquery registers the variant query HTTP API.
package variantquery

import (
	"context"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/variant/clickhouse"
	"github.com/gin-gonic/gin"
)

var safeToken = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// Service is the storage-agnostic dependency required by the handler.
type Service interface {
	Query(context.Context, clickhouse.QueryFilters) (clickhouse.QueryResult, error)
}

// Register mounts GET /variants on the given /v1 group.
func Register(g *gin.RouterGroup, svc Service) {
	if svc == nil {
		g.GET("/variants", handleNoService)
		return
	}
	g.GET("/variants", handleQuery(svc))
}

func handleNoService(c *gin.Context) {
	problem.ServiceUnavailable(c, "Variant query service is not available; set CLICKHOUSE_DSN or CLICKHOUSE_* connection settings.")
}

type response struct {
	Data       []variantOut `json:"data"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	Total      int64        `json:"total"`
	HasNext    bool         `json:"has_next"`
	LatencyMS  int64        `json:"latency_ms"`
	AppliedFor filtersOut   `json:"filters"`
}

type filtersOut struct {
	Chromosome  string  `json:"chromosome,omitempty"`
	PositionMin *uint32 `json:"position_min,omitempty"`
	PositionMax *uint32 `json:"position_max,omitempty"`
	Gene        string  `json:"gene,omitempty"`
}

type variantOut struct {
	Chromosome string   `json:"chromosome"`
	Position   uint32   `json:"position"`
	Ref        string   `json:"ref"`
	Alt        string   `json:"alt"`
	Qual       *float64 `json:"qual,omitempty"`
	Filter     string   `json:"filter"`
	Info       string   `json:"info"`
	Gene       string   `json:"gene,omitempty"`
}

func handleQuery(svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		filters, errs := parseFilters(c)
		if len(errs) > 0 {
			problem.Validation(c, "one or more fields failed validation", errs)
			return
		}
		start := time.Now()
		res, err := svc.Query(c.Request.Context(), filters)
		if err != nil {
			_ = c.Error(err)
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not query variants",
			})
			return
		}
		out := make([]variantOut, 0, len(res.Rows))
		for _, r := range res.Rows {
			out = append(out, variantOut{
				Chromosome: r.Chromosome,
				Position:   r.Position,
				Ref:        r.Ref,
				Alt:        r.Alt,
				Qual:       r.Qual,
				Filter:     r.Filter,
				Info:       r.Info,
				Gene:       r.Gene,
			})
		}
		c.JSON(http.StatusOK, response{
			Data:      out,
			Page:      res.Page,
			PageSize:  res.PageSize,
			Total:     res.Total,
			HasNext:   res.HasNext,
			LatencyMS: time.Since(start).Milliseconds(),
			AppliedFor: filtersOut{
				Chromosome:  filters.Chromosome,
				PositionMin: filters.PositionMin,
				PositionMax: filters.PositionMax,
				Gene:        filters.Gene,
			},
		})
	}
}

func parseFilters(c *gin.Context) (clickhouse.QueryFilters, []problem.FieldError) {
	var errs []problem.FieldError
	chromosome := strings.TrimSpace(c.Query("chromosome"))
	if chromosome != "" && !safeToken.MatchString(chromosome) {
		errs = append(errs, problem.FieldError{Field: "chromosome", Message: "invalid"})
	}

	gene := strings.TrimSpace(c.Query("gene"))
	if gene != "" && !safeToken.MatchString(gene) {
		errs = append(errs, problem.FieldError{Field: "gene", Message: "invalid"})
	}

	posMin, err := parseOptUint32(c.Query("position_min"))
	if err != nil {
		errs = append(errs, problem.FieldError{Field: "position_min", Message: "invalid"})
	}
	posMax, err := parseOptUint32(c.Query("position_max"))
	if err != nil {
		errs = append(errs, problem.FieldError{Field: "position_max", Message: "invalid"})
	}
	if posMin != nil && posMax != nil && *posMin > *posMax {
		errs = append(errs, problem.FieldError{Field: "position_range", Message: "invalid"})
	}

	page, err := parsePositiveInt(c.Query("page"), 1)
	if err != nil {
		errs = append(errs, problem.FieldError{Field: "page", Message: "invalid"})
	}
	pageSize, err := parsePositiveInt(c.Query("page_size"), 50)
	if err != nil {
		errs = append(errs, problem.FieldError{Field: "page_size", Message: "invalid"})
	}
	if pageSize > 200 {
		errs = append(errs, problem.FieldError{Field: "page_size", Message: "max_200"})
	}

	return clickhouse.QueryFilters{
		Chromosome:  chromosome,
		PositionMin: posMin,
		PositionMax: posMax,
		Gene:        gene,
		Page:        page,
		PageSize:    pageSize,
	}, errs
}

func parsePositiveInt(raw string, def int) (int, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, strconv.ErrSyntax
	}
	return n, nil
}

func parseOptUint32(raw string) (*uint32, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, nil
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return nil, err
	}
	u := uint32(n)
	return &u, nil
}
