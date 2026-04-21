// Package mlimpact registers ML baseline APIs for variant impact scoring.
package mlimpact

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/ml/impact"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const stageImpactScored = "impact_scored"

// Service is the scoring dependency required by the handler.
type Service interface {
	Train(ctx context.Context, samples []impact.TrainSample) (impact.Metadata, error)
	Predict(ctx context.Context, in impact.PredictInput) (impact.PredictResult, error)
}

type trainRequest struct {
	Samples []impact.TrainSample `json:"samples"`
}

type trainResponse struct {
	Model impact.Metadata `json:"model"`
}

type predictRequest struct {
	Variant impact.PredictInput `json:"variant"`
}

type predictResponse struct {
	JobID       string             `json:"job_id"`
	Score       float64            `json:"score"`
	Class       string             `json:"class"`
	LatencyMS   int64              `json:"latency_ms"`
	Model       impact.Metadata    `json:"model"`
	Features    map[string]float64 `json:"features"`
	PredictedAt string             `json:"predicted_at"`
}

// Register mounts ML impact endpoints.
func Register(g *gin.RouterGroup, repo job.Repository, svc Service) {
	if repo == nil || svc == nil {
		g.POST("/impact/train", handleUnavailable)
		g.POST("/impact/:job_id/predict", handleUnavailable)
		return
	}
	g.POST("/impact/train", handleTrain(svc))
	g.POST("/impact/:job_id/predict", handlePredict(repo, svc))
}

func handleUnavailable(c *gin.Context) {
	problem.ServiceUnavailable(c, "ML impact service is not available; requires job persistence and scoring service.")
}

func handleTrain(svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req trainRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			problem.MalformedJSON(c, "invalid JSON body")
			return
		}
		if len(req.Samples) == 0 {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "samples", Message: "required"},
			})
			return
		}
		meta, err := svc.Train(c.Request.Context(), req.Samples)
		if err != nil {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "samples", Message: "invalid_training_set"},
			})
			return
		}
		c.JSON(http.StatusOK, trainResponse{Model: meta})
	}
}

func handlePredict(repo job.Repository, svc Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID, err := uuid.Parse(strings.TrimSpace(c.Param("job_id")))
		if err != nil {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "job_id", Message: "invalid_uuid"},
			})
			return
		}
		var req predictRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			problem.MalformedJSON(c, "invalid JSON body")
			return
		}
		result, err := svc.Predict(c.Request.Context(), req.Variant)
		if errors.Is(err, impact.ErrModelNotTrained) {
			problem.JSON(c, http.StatusConflict, problem.Problem{
				Type:   problem.TypeValidationError,
				Title:  "Conflict",
				Status: http.StatusConflict,
				Detail: "model must be trained before prediction",
			})
			return
		}
		if err != nil {
			problem.Validation(c, "one or more fields failed validation", []problem.FieldError{
				{Field: "variant", Message: "invalid"},
			})
			return
		}
		existing, err := repo.GetByID(c.Request.Context(), jobID)
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
		predictedAt := time.Now().UTC().Format(time.RFC3339Nano)
		ref, err := json.Marshal(map[string]any{
			"kind":             "impact_prediction_v1",
			"score":            result.Score,
			"class":            result.Class,
			"latency_ms":       result.LatencyMS,
			"predicted_at":     predictedAt,
			"model":            result.Metadata,
			"feature_vector":   result.Features,
			"feature_version":  result.Metadata.FeatureVersion,
			"model_version":    result.Metadata.ModelVersion,
			"training_dataset": result.Metadata.DatasetHash,
		})
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not build prediction payload",
			})
			return
		}
		if _, err := repo.Update(c.Request.Context(), jobID, job.UpdateParams{
			Status:    existing.Status,
			Stage:     stageImpactScored,
			OutputRef: ref,
		}); err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not persist prediction output",
			})
			return
		}
		c.JSON(http.StatusOK, predictResponse{
			JobID:       jobID.String(),
			Score:       result.Score,
			Class:       result.Class,
			LatencyMS:   result.LatencyMS,
			Model:       result.Metadata,
			Features:    result.Features,
			PredictedAt: predictedAt,
		})
	}
}
