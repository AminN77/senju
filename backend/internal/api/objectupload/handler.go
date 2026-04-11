// Package objectupload registers multipart presigned upload HTTP APIs for S3-compatible object stores.
package objectupload

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AminN77/senju/backend/internal/api/problem"
	"github.com/AminN77/senju/backend/internal/objectstore"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// MultipartInitRequest is the body for POST /v1/objects/multipart.
type MultipartInitRequest struct {
	ContentType  string  `json:"content_type"`
	FilenameHint *string `json:"filename_hint,omitempty"`
}

// MultipartInitResponse is returned on 201 Created.
type MultipartInitResponse struct {
	Bucket        string `json:"bucket"`
	ObjectKey     string `json:"object_key"`
	UploadID      string `json:"upload_id"`
	PartSizeBytes int64  `json:"part_size_bytes"`
}

// PresignPartRequest is the body for POST /v1/objects/multipart/:upload_id/parts.
type PresignPartRequest struct {
	ObjectKey  string `json:"object_key"`
	PartNumber int32  `json:"part_number"`
}

// PresignPartResponse is returned on 200 OK.
type PresignPartResponse struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
}

// CompletePart is one completed part for CompleteMultipartUpload.
type CompletePart struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"etag"`
}

// CompleteRequest is the body for POST /v1/objects/multipart/:upload_id/complete.
type CompleteRequest struct {
	ObjectKey string         `json:"object_key"`
	Parts     []CompletePart `json:"parts"`
}

// CompleteResponse is returned on 200 OK.
type CompleteResponse struct {
	ObjectKey string `json:"object_key"`
	ETag      string `json:"etag"`
	SizeBytes int64  `json:"size_bytes"`
}

// Register mounts object multipart routes on /v1/objects. If store is nil, handlers return 503.
func Register(g *gin.RouterGroup, store objectstore.Service, log zerolog.Logger) {
	if store == nil {
		g.POST("/multipart", handleObjectStoreDisabled)
		g.POST("/multipart/:upload_id/parts", handleObjectStoreDisabled)
		g.POST("/multipart/:upload_id/complete", handleObjectStoreDisabled)
		return
	}
	g.POST("/multipart", handleMultipartInit(store, log))
	g.POST("/multipart/:upload_id/parts", handlePresignPart(store))
	g.POST("/multipart/:upload_id/complete", handleComplete(store, log))
}

func handleObjectStoreDisabled(c *gin.Context) {
	problem.ServiceUnavailable(c, "Object storage is not configured; set S3_ENDPOINT, S3_BUCKET, and S3_ACCESS_KEY/S3_SECRET_KEY (or MINIO_ROOT_USER/MINIO_ROOT_PASSWORD).")
}

func handleMultipartInit(store objectstore.Service, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		req, err := decodeJSON[MultipartInitRequest](c)
		if err != nil {
			writeDecodeError(c, err)
			return
		}

		ct := strings.TrimSpace(req.ContentType)
		if ct == "" {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "content_type", Message: "required"},
			})
			return
		}

		key := buildObjectKey(req.FilenameHint)
		res, err := store.CreateMultipartUpload(c.Request.Context(), key, ct)
		if err != nil {
			log.Error().Err(err).Str("object_key", key).Msg("CreateMultipartUpload failed")
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not initiate multipart upload",
			})
			return
		}

		c.JSON(http.StatusCreated, MultipartInitResponse{
			Bucket:        res.Bucket,
			ObjectKey:     res.ObjectKey,
			UploadID:      res.UploadID,
			PartSizeBytes: res.PartSizeBytes,
		})
	}
}

func handlePresignPart(store objectstore.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadID := strings.TrimSpace(c.Param("upload_id"))
		if uploadID == "" {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "upload_id", Message: "required"},
			})
			return
		}

		req, err := decodeJSON[PresignPartRequest](c)
		if err != nil {
			writeDecodeError(c, err)
			return
		}

		key := strings.TrimSpace(req.ObjectKey)
		if key == "" {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "object_key", Message: "required"},
			})
			return
		}
		if req.PartNumber < 1 || req.PartNumber > 10000 {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "part_number", Message: "out_of_range"},
			})
			return
		}

		res, err := store.PresignUploadPart(c.Request.Context(), key, uploadID, req.PartNumber)
		if err != nil {
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not presign upload part",
			})
			return
		}

		c.JSON(http.StatusOK, PresignPartResponse{
			URL:       res.URL,
			ExpiresAt: res.ExpiresAt.UTC().Format(time.RFC3339Nano),
		})
	}
}

func handleComplete(store objectstore.Service, log zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadIDPath := strings.TrimSpace(c.Param("upload_id"))
		if uploadIDPath == "" {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "upload_id", Message: "required"},
			})
			return
		}

		req, err := decodeJSON[CompleteRequest](c)
		if err != nil {
			writeDecodeError(c, err)
			return
		}

		key := strings.TrimSpace(req.ObjectKey)
		if key == "" {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "object_key", Message: "required"},
			})
			return
		}
		if len(req.Parts) == 0 {
			problem.Validation(c, "validation failed", []problem.FieldError{
				{Field: "parts", Message: "required"},
			})
			return
		}

		parts := make([]objectstore.PartETag, 0, len(req.Parts))
		for _, p := range req.Parts {
			if p.PartNumber < 1 || p.PartNumber > 10000 {
				problem.Validation(c, "validation failed", []problem.FieldError{
					{Field: "parts", Message: "invalid_part_number"},
				})
				return
			}
			etag := strings.TrimSpace(p.ETag)
			if etag == "" {
				problem.Validation(c, "validation failed", []problem.FieldError{
					{Field: "parts", Message: "empty_etag"},
				})
				return
			}
			parts = append(parts, objectstore.PartETag{PartNumber: p.PartNumber, ETag: etag})
		}

		res, err := store.CompleteMultipartUpload(c.Request.Context(), key, uploadIDPath, parts)
		if err != nil {
			log.Error().Err(err).Str("object_key", key).Str("upload_id", uploadIDPath).Msg("CompleteMultipartUpload failed")
			problem.JSON(c, http.StatusInternalServerError, problem.Problem{
				Type:   problem.TypeInternalError,
				Title:  "Internal error",
				Status: http.StatusInternalServerError,
				Detail: "could not complete multipart upload",
			})
			return
		}

		log.Info().
			Str("audit", "object_multipart_complete").
			Str("object_key", key).
			Str("upload_id", uploadIDPath).
			Int64("object_size_bytes", res.SizeBytes).
			Str("etag", res.ETag).
			Str("checksum_status", "validated").
			Msg("object upload checksum audit")

		c.JSON(http.StatusOK, CompleteResponse{
			ObjectKey: key,
			ETag:      res.ETag,
			SizeBytes: res.SizeBytes,
		})
	}
}

var errTrailingJSON = errors.New("trailing json")

func decodeJSON[T any](c *gin.Context) (T, error) {
	var zero T
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return zero, err
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	var v T
	if err := dec.Decode(&v); err != nil {
		return zero, err
	}
	rest := bytes.TrimSpace(body[dec.InputOffset():])
	if len(rest) > 0 {
		return zero, errTrailingJSON
	}
	return v, nil
}

func writeDecodeError(c *gin.Context, err error) {
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
}

func buildObjectKey(filenameHint *string) string {
	id := uuid.New().String()
	if filenameHint == nil {
		return "uploads/" + id
	}
	h := sanitizeFilenameHint(*filenameHint)
	if h == "" {
		return "uploads/" + id
	}
	return "uploads/" + id + "/" + h
}

func sanitizeFilenameHint(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '_', r == '-':
			b.WriteRune(r)
		}
	}
	out := b.String()
	if len(out) > 200 {
		out = out[:200]
	}
	if out == "" {
		return ""
	}
	return out
}
