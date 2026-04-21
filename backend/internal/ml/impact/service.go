// Package impact provides a baseline variant impact scoring service.
package impact

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// FeatureVersion identifies the current feature extraction contract.
	FeatureVersion = "impact_features_v1"
)

var (
	// ErrModelNotTrained is returned when prediction is requested before training.
	ErrModelNotTrained = errors.New("impact: model not trained")
	// ErrEmptyTrainingSet is returned when no training samples are supplied.
	ErrEmptyTrainingSet = errors.New("impact: empty training set")
)

// TrainSample is one labeled variant sample used for training.
type TrainSample struct {
	Chromosome string  `json:"chromosome"`
	Position   uint32  `json:"position"`
	Ref        string  `json:"ref"`
	Alt        string  `json:"alt"`
	Qual       float64 `json:"qual"`
	Filter     string  `json:"filter"`
	Gene       string  `json:"gene,omitempty"`
	Label      int     `json:"label"`
}

// PredictInput is one unlabeled variant used for scoring.
type PredictInput struct {
	Chromosome string  `json:"chromosome"`
	Position   uint32  `json:"position"`
	Ref        string  `json:"ref"`
	Alt        string  `json:"alt"`
	Qual       float64 `json:"qual"`
	Filter     string  `json:"filter"`
	Gene       string  `json:"gene,omitempty"`
}

// Metadata captures reproducibility metadata for a trained model.
type Metadata struct {
	DatasetHash    string `json:"dataset_hash"`
	FeatureVersion string `json:"feature_version"`
	ModelVersion   string `json:"model_version"`
	TrainedAt      string `json:"trained_at"`
	SampleCount    int    `json:"sample_count"`
}

// PredictResult is one inference result.
type PredictResult struct {
	Score     float64            `json:"score"`
	Class     string             `json:"class"`
	LatencyMS int64              `json:"latency_ms"`
	Features  map[string]float64 `json:"features"`
	Metadata  Metadata           `json:"metadata"`
}

// Service implements training and online prediction for baseline impact scoring.
type Service struct {
	mu       sync.RWMutex
	trained  bool
	weights  []float64
	bias     float64
	metadata Metadata
}

// NewService constructs a baseline impact scoring service.
func NewService() *Service {
	return &Service{
		weights: make([]float64, 5),
	}
}

// Train builds a baseline linear classifier and stores model metadata.
func (s *Service) Train(_ context.Context, samples []TrainSample) (Metadata, error) {
	if len(samples) == 0 {
		return Metadata{}, ErrEmptyTrainingSet
	}
	dh, err := datasetHash(samples)
	if err != nil {
		return Metadata{}, fmt.Errorf("impact: dataset hash: %w", err)
	}
	pos := make([]float64, 5)
	neg := make([]float64, 5)
	var posN, negN float64
	for _, sample := range samples {
		fv, err := extractFeatures(sample.Chromosome, sample.Ref, sample.Alt, sample.Qual, sample.Filter, sample.Gene)
		if err != nil {
			return Metadata{}, err
		}
		switch sample.Label {
		case 1:
			for i := range fv {
				pos[i] += fv[i]
			}
			posN++
		case 0:
			for i := range fv {
				neg[i] += fv[i]
			}
			negN++
		default:
			return Metadata{}, fmt.Errorf("impact: label must be 0 or 1")
		}
	}
	if posN == 0 || negN == 0 {
		return Metadata{}, fmt.Errorf("impact: training requires both positive and negative labels")
	}
	for i := range pos {
		pos[i] /= posN
		neg[i] /= negN
	}
	weights := make([]float64, 5)
	for i := range weights {
		weights[i] = pos[i] - neg[i]
	}
	bias := 0.0
	for i := range weights {
		bias += -0.5 * weights[i] * (pos[i] + neg[i])
	}
	meta := Metadata{
		DatasetHash:    dh,
		FeatureVersion: FeatureVersion,
		ModelVersion:   "impact-baseline-" + strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", ""),
		TrainedAt:      time.Now().UTC().Format(time.RFC3339Nano),
		SampleCount:    len(samples),
	}
	s.mu.Lock()
	s.weights = weights
	s.bias = bias
	s.metadata = meta
	s.trained = true
	s.mu.Unlock()
	return meta, nil
}

// Predict scores one variant sample.
func (s *Service) Predict(_ context.Context, in PredictInput) (PredictResult, error) {
	start := time.Now()
	fv, err := extractFeatures(in.Chromosome, in.Ref, in.Alt, in.Qual, in.Filter, in.Gene)
	if err != nil {
		return PredictResult{}, err
	}
	s.mu.RLock()
	if !s.trained {
		s.mu.RUnlock()
		return PredictResult{}, ErrModelNotTrained
	}
	weights := append([]float64(nil), s.weights...)
	bias := s.bias
	meta := s.metadata
	s.mu.RUnlock()
	logit := bias
	for i := range weights {
		logit += weights[i] * fv[i]
	}
	score := 1 / (1 + math.Exp(-logit))
	class := "benign"
	if score >= 0.5 {
		class = "deleterious"
	}
	return PredictResult{
		Score:     score,
		Class:     class,
		LatencyMS: time.Since(start).Milliseconds(),
		Features: map[string]float64{
			"qual_scaled":     fv[0],
			"alt_len_scaled":  fv[1],
			"ref_len_scaled":  fv[2],
			"is_pass":         fv[3],
			"has_gene_symbol": fv[4],
		},
		Metadata: meta,
	}, nil
}

func extractFeatures(chromosome, ref, alt string, qual float64, filter, gene string) ([]float64, error) {
	if strings.TrimSpace(chromosome) == "" {
		return nil, fmt.Errorf("impact: chromosome is required")
	}
	if strings.TrimSpace(ref) == "" || strings.TrimSpace(alt) == "" {
		return nil, fmt.Errorf("impact: ref and alt are required")
	}
	if qual < 0 {
		return nil, fmt.Errorf("impact: qual must be >= 0")
	}
	isPass := 0.0
	if strings.EqualFold(strings.TrimSpace(filter), "PASS") {
		isPass = 1.0
	}
	hasGene := 0.0
	if strings.TrimSpace(gene) != "" {
		hasGene = 1.0
	}
	return []float64{
		math.Min(qual/100.0, 1.0),
		math.Min(float64(len(strings.TrimSpace(alt)))/10.0, 1.0),
		math.Min(float64(len(strings.TrimSpace(ref)))/10.0, 1.0),
		isPass,
		hasGene,
	}, nil
}

func datasetHash(samples []TrainSample) (string, error) {
	cp := append([]TrainSample(nil), samples...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].Chromosome != cp[j].Chromosome {
			return cp[i].Chromosome < cp[j].Chromosome
		}
		if cp[i].Position != cp[j].Position {
			return cp[i].Position < cp[j].Position
		}
		if cp[i].Ref != cp[j].Ref {
			return cp[i].Ref < cp[j].Ref
		}
		if cp[i].Alt != cp[j].Alt {
			return cp[i].Alt < cp[j].Alt
		}
		return cp[i].Label < cp[j].Label
	})
	raw, err := json.Marshal(cp)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
