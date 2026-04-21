// Package fastqupload registers the FASTQ upload metadata HTTP API.
package fastqupload

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/fastq"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// StageFastqUploadQueued is the initial pipeline stage for a FASTQ metadata job row.
const StageFastqUploadQueued = "fastq_upload_queued"

// StageFastqValidated indicates FASTQ stream format passed validation.
const StageFastqValidated = "fastq_validated"

// StageFastqValidationFailed indicates FASTQ stream format failed validation.
const StageFastqValidationFailed = "fastq_validation_failed"

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

// ValidateResponse is returned by FASTQ streaming format validation.
type ValidateResponse struct {
	Valid         bool   `json:"valid"`
	Records       int64  `json:"records"`
	FailureReason string `json:"failure_reason,omitempty"`
	FailureRecord int64  `json:"failure_record,omitempty"`
}

// Register mounts POST /fastq-upload/metadata on the given /v1/jobs group.
// If repo is nil, the handler returns 503 (database not configured).
func Register(g *gin.RouterGroup, repo job.Repository) {
	if repo == nil {
		g.POST("/fastq-upload/metadata", handleNoDatabase)
		g.POST("/fastq-upload/:job_id/validate", handleNoDatabase)
		return
	}
	g.POST("/fastq-upload/metadata", handleCreate(repo))
	g.POST("/fastq-upload/:job_id/validate", handleValidate(repo))
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
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return Request{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	var req Request
	if err := dec.Decode(&req); err != nil {
		return req, err
	}
	// Decoder.More() is for arrays/objects, not extra top-level values; use input offset.
	rest := bytes.TrimSpace(body[dec.InputOffset():])
	if len(rest) > 0 {
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
		normalized := strings.TrimSpace(pair.raw)
		if normalized == "" {
			errs = append(errs, problem.FieldError{Field: pair.field, Message: "required"})
			continue
		}
		if _, err := validateHTTPOrObjectURI(normalized); err != nil {
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

func handleValidate(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobIDRaw := strings.TrimSpace(c.Param("job_id"))
		jobID, err := uuid.Parse(jobIDRaw)
		if err != nil {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "job_id", Message: "invalid_uuid"},
			})
			return
		}

		res, err := fastq.ValidateStream(c.Request.Body)
		if err != nil {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "fastq", Message: "could_not_validate_stream"},
			})
			return
		}

		stage := StageFastqValidated
		status := job.StatusSucceeded
		if !res.Valid {
			stage = StageFastqValidationFailed
			status = job.StatusFailed
		}

		payload, err := json.Marshal(map[string]any{
			"kind":           "fastq_validation_v1",
			"valid":          res.Valid,
			"records":        res.Records,
			"failure_reason": res.FailureReason,
			"failure_record": res.FailureRecord,
			"validated_at":   time.Now().UTC().Format(time.RFC3339Nano),
		})
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not build validation payload",
			})
			return
		}

		_, err = repo.Update(c.Request.Context(), jobID, job.UpdateParams{
			Status:    status,
			Stage:     stage,
			OutputRef: payload,
		})
		if errors.Is(err, job.ErrNotFound) {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "job_id", Message: "not_found"},
			})
			return
		}
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not persist validation result",
			})
			return
		}

		c.JSON(http.StatusOK, ValidateResponse{
			Valid:         res.Valid,
			Records:       res.Records,
			FailureReason: res.FailureReason,
			FailureRecord: res.FailureRecord,
		})
	}
}
