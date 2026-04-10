// Package fastqupload registers the FASTQ upload metadata HTTP API.
package fastqupload

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/gin-gonic/gin"
)

// StageFastqUploadQueued is the initial pipeline stage for a FASTQ metadata job row.
const StageFastqUploadQueued = "fastq_upload_queued"

// Request is the JSON body for POST .../fastq-upload/metadata (paired-end MVP).
type Request struct {
	SampleID  string  `json:"sample_id"`
	R1URI     string  `json:"r1_uri"`
	R2URI     string  `json:"r2_uri"`
	LibraryID *string `json:"library_id,omitempty"`
	Platform  *string `json:"platform,omitempty"`
	Notes     *string `json:"notes,omitempty"`
}

// CreateResponse is returned on 201 Created.
type CreateResponse struct {
	JobID string `json:"job_id"`
}

// Register mounts POST /fastq-upload/metadata on the given /v1/jobs group.
// If repo is nil, the handler returns 503 (database not configured).
func Register(g *gin.RouterGroup, repo job.Repository) {
	if repo == nil {
		g.POST("/fastq-upload/metadata", handleNoDatabase)
		return
	}
	g.POST("/fastq-upload/metadata", handleCreate(repo))
}

func handleNoDatabase(c *gin.Context) {
	problem.ServiceUnavailable(c, "Job persistence is not available; set POSTGRES_DSN or POSTGRES_HOST with credentials.")
}

func handleCreate(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeRequest(c)
		if err != nil {
			var synErr *json.SyntaxError
			if errors.As(err, &synErr) {
				problem.MalformedJSON(c, "invalid JSON syntax")
				return
			}
			if errors.Is(err, errTrailingJSON) {
				problem.MalformedJSON(c, "trailing JSON content")
				return
			}
			if strings.Contains(err.Error(), "unknown field") {
				problem.MalformedJSON(c, "unknown JSON field")
				return
			}
			problem.MalformedJSON(c, "could not read request body")
			return
		}

		fieldErrs := validateRequest(req)
		if len(fieldErrs) > 0 {
			problem.Validation(c, "one or more fields failed validation", fieldErrs)
			return
		}

		input, err := canonicalInputJSON(req)
		if err != nil {
			problem.MalformedJSON(c, "could not build job payload")
			return
		}

		j, err := repo.Create(c.Request.Context(), job.CreateParams{
			Status:    job.StatusPending,
			Stage:     StageFastqUploadQueued,
			InputRef:  input,
			OutputRef: nil,
		})
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not persist job",
			})
			return
		}

		c.JSON(http.StatusCreated, CreateResponse{JobID: j.ID.String()})
	}
}

var errTrailingJSON = errors.New("trailing json")

func decodeRequest(c *gin.Context) (Request, error) {
	var req Request
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return req, err
	}
	if dec.More() {
		return req, errTrailingJSON
	}
	return req, nil
}

func validateRequest(req Request) []problem.FieldError {
	var errs []problem.FieldError
	sid := strings.TrimSpace(req.SampleID)
	if sid == "" {
		errs = append(errs, problem.FieldError{Field: "sample_id", Message: "required"})
	}
	for _, pair := range []struct {
		field string
		raw   string
	}{
		{"r1_uri", req.R1URI},
		{"r2_uri", req.R2URI},
	} {
		if strings.TrimSpace(pair.raw) == "" {
			errs = append(errs, problem.FieldError{Field: pair.field, Message: "required"})
			continue
		}
		if _, err := validateHTTPOrObjectURI(pair.raw); err != nil {
			errs = append(errs, problem.FieldError{Field: pair.field, Message: "invalid_uri"})
		}
	}
	return errs
}

func validateHTTPOrObjectURI(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("missing scheme or host")
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https":
		return u, nil
	case "s3", "s3a", "gs":
		key := strings.Trim(strings.TrimPrefix(u.Path, "/"), "/")
		if key == "" {
			return nil, errors.New("object key required")
		}
		return u, nil
	default:
		return nil, errors.New("unsupported scheme")
	}
}

func canonicalInputJSON(req Request) (json.RawMessage, error) {
	canon := map[string]any{
		"kind":      "fastq_upload_v1",
		"sample_id": strings.TrimSpace(req.SampleID),
		"r1_uri":    strings.TrimSpace(req.R1URI),
		"r2_uri":    strings.TrimSpace(req.R2URI),
	}
	if req.LibraryID != nil && strings.TrimSpace(*req.LibraryID) != "" {
		canon["library_id"] = strings.TrimSpace(*req.LibraryID)
	}
	if req.Platform != nil && strings.TrimSpace(*req.Platform) != "" {
		canon["platform"] = strings.TrimSpace(*req.Platform)
	}
	if req.Notes != nil && strings.TrimSpace(*req.Notes) != "" {
		canon["notes"] = strings.TrimSpace(*req.Notes)
	}
	return json.Marshal(canon)
}
