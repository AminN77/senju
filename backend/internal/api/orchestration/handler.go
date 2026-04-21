// Package orchestration registers pipeline orchestration HTTP APIs.
package orchestration

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	stagePipelineQueued    = "pipeline_queued"
	stagePipelineRunning   = "pipeline_running"
	stagePipelineSucceeded = "pipeline_succeeded"
	stagePipelineFailed    = "pipeline_failed"
)

type createRequest struct {
	SampleID  string `json:"sample_id"`
	R1URI     string `json:"r1_uri"`
	R2URI     string `json:"r2_uri"`
	ForceFail bool   `json:"force_fail,omitempty"`
}

type createResponse struct {
	JobID string `json:"job_id"`
}

type runResponse struct {
	JobID  string     `json:"job_id"`
	Status job.Status `json:"status"`
	Stage  string     `json:"stage"`
}

type statusResponse struct {
	JobID       string     `json:"job_id"`
	Status      job.Status `json:"status"`
	Stage       string     `json:"stage"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Register mounts orchestration endpoints on the given /v1/jobs group.
func Register(g *gin.RouterGroup, repo job.Repository) {
	if repo == nil {
		g.POST("/pipeline", handleNoDatabase)
		g.POST("/:job_id/run", handleNoDatabase)
		g.GET("/:job_id/status", handleNoDatabase)
		g.GET("/:job_id/outputs", handleNoDatabase)
		return
	}
	g.POST("/pipeline", handleCreate(repo))
	g.POST("/:job_id/run", handleRun(repo))
	g.GET("/:job_id/status", handleStatus(repo))
	g.GET("/:job_id/outputs", handleOutputs(repo))
}

func handleNoDatabase(c *gin.Context) {
	problem.ServiceUnavailable(c, "Job persistence is not available; set POSTGRES_DSN or POSTGRES_HOST with credentials.")
}

func handleCreate(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			problem.MalformedJSON(c, "invalid JSON body")
			return
		}
		fieldErrs := validateCreateRequest(req)
		if len(fieldErrs) > 0 {
			problem.Validation(c, "one or more fields failed validation", fieldErrs)
			return
		}
		inputRef, err := json.Marshal(map[string]any{
			"kind":       "pipeline_job_v1",
			"sample_id":  strings.TrimSpace(req.SampleID),
			"r1_uri":     strings.TrimSpace(req.R1URI),
			"r2_uri":     strings.TrimSpace(req.R2URI),
			"force_fail": req.ForceFail,
		})
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not build job payload",
			})
			return
		}
		created, err := repo.Create(c.Request.Context(), job.CreateParams{
			Status:   job.StatusPending,
			Stage:    stagePipelineQueued,
			InputRef: inputRef,
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
		c.JSON(http.StatusCreated, createResponse{JobID: created.ID.String()})
	}
}

func handleRun(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, ok := parseJobID(c)
		if !ok {
			return
		}
		j, err := repo.GetByID(c.Request.Context(), jobID)
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
				Detail: "could not fetch job",
			})
			return
		}
		if j.Status != job.StatusPending || j.Stage != stagePipelineQueued {
			problem.JSON(c, http.StatusConflict, problem.Problem{
				Type:   problem.TypeValidationError,
				Title:  "Conflict",
				Status: http.StatusConflict,
				Detail: "job is not in runnable state",
			})
			return
		}
		now := time.Now().UTC()
		startedAudit, err := buildAuditOutput("running", "", now, nil)
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not build run metadata",
			})
			return
		}
		if _, err := repo.Update(c.Request.Context(), jobID, job.UpdateParams{
			Status:    job.StatusRunning,
			Stage:     stagePipelineRunning,
			StartedAt: &now,
			OutputRef: startedAudit,
		}); err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not persist running state",
			})
			return
		}
		forceFail := false
		var payload createRequest
		if err := json.Unmarshal(j.InputRef, &payload); err == nil {
			forceFail = payload.ForceFail
		}
		completedAt := time.Now().UTC()
		result := "succeeded"
		errorMsg := ""
		nextStatus := job.StatusSucceeded
		nextStage := stagePipelineSucceeded
		if forceFail {
			result = "failed"
			errorMsg = "forced_failure"
			nextStatus = job.StatusFailed
			nextStage = stagePipelineFailed
		}
		finishedAudit, err := buildAuditOutput(result, errorMsg, now, &completedAt)
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not build completion metadata",
			})
			return
		}
		if _, err := repo.Update(c.Request.Context(), jobID, job.UpdateParams{
			Status:      nextStatus,
			Stage:       nextStage,
			CompletedAt: &completedAt,
			OutputRef:   finishedAudit,
		}); err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not persist completion state",
			})
			return
		}
		c.JSON(http.StatusOK, runResponse{
			JobID:  jobID.String(),
			Status: nextStatus,
			Stage:  nextStage,
		})
	}
}

func handleStatus(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, ok := parseJobID(c)
		if !ok {
			return
		}
		j, err := repo.GetByID(c.Request.Context(), jobID)
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
				Detail: "could not fetch job",
			})
			return
		}
		c.JSON(http.StatusOK, statusResponse{
			JobID:       j.ID.String(),
			Status:      j.Status,
			Stage:       j.Stage,
			StartedAt:   j.StartedAt,
			CompletedAt: j.CompletedAt,
			UpdatedAt:   j.UpdatedAt,
		})
	}
}

func handleOutputs(repo job.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, ok := parseJobID(c)
		if !ok {
			return
		}
		j, err := repo.GetByID(c.Request.Context(), jobID)
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
				Detail: "could not fetch job",
			})
			return
		}
		if len(j.OutputRef) == 0 {
			c.JSON(http.StatusOK, gin.H{"job_id": jobID.String(), "output_ref": json.RawMessage("null")})
			return
		}
		c.JSON(http.StatusOK, gin.H{"job_id": jobID.String(), "output_ref": j.OutputRef})
	}
}

func parseJobID(c *gin.Context) (uuid.UUID, bool) {
	jobID, err := uuid.Parse(strings.TrimSpace(c.Param("job_id")))
	if err != nil {
		problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
			{Field: "job_id", Message: "invalid_uuid"},
		})
		return uuid.Nil, false
	}
	return jobID, true
}

func validateCreateRequest(req createRequest) []problem.FieldError {
	var errs []problem.FieldError
	if strings.TrimSpace(req.SampleID) == "" {
		errs = append(errs, problem.FieldError{Field: "sample_id", Message: "required"})
	}
	if strings.TrimSpace(req.R1URI) == "" {
		errs = append(errs, problem.FieldError{Field: "r1_uri", Message: "required"})
	}
	if strings.TrimSpace(req.R2URI) == "" {
		errs = append(errs, problem.FieldError{Field: "r2_uri", Message: "required"})
	}
	return errs
}

func buildAuditOutput(result, errorMsg string, startedAt time.Time, completedAt *time.Time) (json.RawMessage, error) {
	transitionLog := []map[string]string{
		{
			"from": "pipeline_queued",
			"to":   "pipeline_running",
			"at":   startedAt.Format(time.RFC3339Nano),
		},
	}
	if completedAt != nil {
		to := stagePipelineSucceeded
		if result == "failed" {
			to = stagePipelineFailed
		}
		transitionLog = append(transitionLog, map[string]string{
			"from": stagePipelineRunning,
			"to":   to,
			"at":   completedAt.Format(time.RFC3339Nano),
		})
	}
	payload := map[string]any{
		"kind":            "pipeline_orchestration_v1",
		"result":          result,
		"started_at":      startedAt.Format(time.RFC3339Nano),
		"transition_log":  transitionLog,
		"audit_generated": time.Now().UTC().Format(time.RFC3339Nano),
	}
	if completedAt != nil {
		payload["completed_at"] = completedAt.Format(time.RFC3339Nano)
	}
	if errorMsg != "" {
		payload["error"] = errorMsg
	}
	return json.Marshal(payload)
}
