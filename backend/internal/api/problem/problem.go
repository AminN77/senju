// Package problem provides RFC 7807-style JSON error bodies for HTTP APIs.
package problem

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

// Well-known problem type URIs (stable identifiers; not necessarily resolvable).
const (
	TypeValidationError = "https://github.com/AminN77/senju/problems/validation-error"
	TypeMalformedJSON   = "https://github.com/AminN77/senju/problems/malformed-json"
	TypeServiceDisabled = "https://github.com/AminN77/senju/problems/service-unavailable"
	TypeInternalError   = "https://github.com/AminN77/senju/problems/internal-error"
)

// Problem is an application/problem+json response (RFC 7807 profile).
type Problem struct {
	Type   string       `json:"type"`
	Title  string       `json:"title"`
	Status int          `json:"status"`
	Detail string       `json:"detail,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

// FieldError is a single validation failure with stable machine-oriented messaging.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// JSON writes a problem+json response with the given HTTP status.
func JSON(c *gin.Context, status int, p Problem) {
	c.Header("Content-Type", "application/problem+json; charset=utf-8")
	c.JSON(status, p)
}

// Validation writes a deterministic validation error (errors sorted by field).
func Validation(c *gin.Context, detail string, errs []FieldError) {
	sort.Slice(errs, func(i, j int) bool { return errs[i].Field < errs[j].Field })
	JSON(c, http.StatusBadRequest, Problem{
		Type:   TypeValidationError,
		Title:  "Validation failed",
		Status: http.StatusBadRequest,
		Detail: detail,
		Errors: errs,
	})
}

// MalformedJSON writes 400 for JSON parse / unknown field issues.
func MalformedJSON(c *gin.Context, detail string) {
	JSON(c, http.StatusBadRequest, Problem{
		Type:   TypeMalformedJSON,
		Title:  "Malformed JSON",
		Status: http.StatusBadRequest,
		Detail: detail,
	})
}

// ServiceUnavailable writes 503 when an optional backend (e.g. Postgres) is not configured.
func ServiceUnavailable(c *gin.Context, detail string) {
	JSON(c, http.StatusServiceUnavailable, Problem{
		Type:   TypeServiceDisabled,
		Title:  "Service unavailable",
		Status: http.StatusServiceUnavailable,
		Detail: detail,
	})
}
