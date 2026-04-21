package impact

import (
	"context"
	"testing"
)

func TestServiceTrainAndPredict(t *testing.T) {
	t.Parallel()
	svc := NewService()
	meta, err := svc.Train(context.Background(), []TrainSample{
		{Chromosome: "chr1", Position: 10, Ref: "A", Alt: "T", Qual: 90, Filter: "PASS", Gene: "TP53", Label: 1},
		{Chromosome: "chr2", Position: 11, Ref: "A", Alt: "G", Qual: 85, Filter: "PASS", Gene: "BRCA1", Label: 1},
		{Chromosome: "chr3", Position: 12, Ref: "A", Alt: "C", Qual: 10, Filter: "q10", Label: 0},
		{Chromosome: "chr4", Position: 13, Ref: "A", Alt: "C", Qual: 12, Filter: "q10", Label: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.DatasetHash == "" || meta.ModelVersion == "" || meta.FeatureVersion != FeatureVersion {
		t.Fatalf("metadata %+v", meta)
	}
	got, err := svc.Predict(context.Background(), PredictInput{
		Chromosome: "chr17", Position: 7579472, Ref: "C", Alt: "T", Qual: 99, Filter: "PASS", Gene: "TP53",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Score <= 0 || got.Score >= 1 {
		t.Fatalf("score %.6f", got.Score)
	}
	if got.Metadata.DatasetHash != meta.DatasetHash {
		t.Fatalf("metadata mismatch %+v %+v", got.Metadata, meta)
	}
}

func TestServicePredictWithoutTraining(t *testing.T) {
	t.Parallel()
	svc := NewService()
	_, err := svc.Predict(context.Background(), PredictInput{
		Chromosome: "chr1", Ref: "A", Alt: "T", Qual: 10, Filter: "PASS",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
