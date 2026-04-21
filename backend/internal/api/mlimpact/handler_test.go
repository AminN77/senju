package mlimpact

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/AminN77/senju/backend/internal/job"
	"github.com/AminN77/senju/backend/internal/job/stub"
	"github.com/AminN77/senju/backend/internal/ml/impact"
	"github.com/gin-gonic/gin"
)

func TestTrainAndPredict_PersistsOutput(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	svc := impact.NewService()
	r := gin.New()
	Register(r.Group("/v1/ml"), repo, svc)
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusSucceeded, Stage: "pipeline_succeeded"})
	if err != nil {
		t.Fatal(err)
	}
	trainBody := `{"samples":[
{"chromosome":"chr1","position":100,"ref":"A","alt":"T","qual":95,"filter":"PASS","gene":"TP53","label":1},
{"chromosome":"chr2","position":101,"ref":"G","alt":"C","qual":88,"filter":"PASS","gene":"BRCA1","label":1},
{"chromosome":"chr3","position":102,"ref":"C","alt":"T","qual":8,"filter":"q10","label":0},
{"chromosome":"chr4","position":103,"ref":"T","alt":"A","qual":12,"filter":"q10","label":0}
]}`
	trainReq := httptest.NewRequest(http.MethodPost, "/v1/ml/impact/train", strings.NewReader(trainBody))
	trainReq.Header.Set("Content-Type", "application/json")
	trainResp := httptest.NewRecorder()
	r.ServeHTTP(trainResp, trainReq)
	if trainResp.Code != http.StatusOK {
		t.Fatalf("train status %d body %s", trainResp.Code, trainResp.Body.String())
	}

	predictBody := `{"variant":{"chromosome":"chr17","position":7579472,"ref":"C","alt":"T","qual":99,"filter":"PASS","gene":"TP53"}}`
	predictReq := httptest.NewRequest(http.MethodPost, "/v1/ml/impact/"+created.ID.String()+"/predict", strings.NewReader(predictBody))
	predictReq.Header.Set("Content-Type", "application/json")
	predictResp := httptest.NewRecorder()
	r.ServeHTTP(predictResp, predictReq)
	if predictResp.Code != http.StatusOK {
		t.Fatalf("predict status %d body %s", predictResp.Code, predictResp.Body.String())
	}
	updated, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Stage != stageImpactScored {
		t.Fatalf("stage %s", updated.Stage)
	}
	if !strings.Contains(string(updated.OutputRef), `"dataset_hash"`) || !strings.Contains(string(updated.OutputRef), `"score"`) {
		t.Fatalf("output_ref %s", updated.OutputRef)
	}
}

func TestPredict_ConflictWhenModelMissing(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	svc := impact.NewService()
	r := gin.New()
	Register(r.Group("/v1/ml"), repo, svc)
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusPending, Stage: "queued"})
	if err != nil {
		t.Fatal(err)
	}
	predictBody := `{"variant":{"chromosome":"chr1","position":1,"ref":"A","alt":"T","qual":10,"filter":"PASS"}}`
	req := httptest.NewRequest(http.MethodPost, "/v1/ml/impact/"+created.ID.String()+"/predict", strings.NewReader(predictBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestPredictP95Under100ms(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)
	repo := stub.New()
	svc := impact.NewService()
	r := gin.New()
	Register(r.Group("/v1/ml"), repo, svc)
	_, err := svc.Train(context.Background(), []impact.TrainSample{
		{Chromosome: "chr1", Position: 1, Ref: "A", Alt: "T", Qual: 90, Filter: "PASS", Gene: "TP53", Label: 1},
		{Chromosome: "chr2", Position: 2, Ref: "A", Alt: "C", Qual: 10, Filter: "q10", Label: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	created, err := repo.Create(context.Background(), job.CreateParams{Status: job.StatusRunning, Stage: "pipeline_running"})
	if err != nil {
		t.Fatal(err)
	}
	const samples = 200
	lat := make([]time.Duration, 0, samples)
	for range samples {
		body := `{"variant":{"chromosome":"chr17","position":7579472,"ref":"C","alt":"T","qual":99,"filter":"PASS","gene":"TP53"}}`
		req := httptest.NewRequest(http.MethodPost, "/v1/ml/impact/"+created.ID.String()+"/predict", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		start := time.Now()
		r.ServeHTTP(w, req)
		lat = append(lat, time.Since(start))
		if w.Code != http.StatusOK {
			t.Fatalf("status %d body %s", w.Code, w.Body.String())
		}
		var out predictResponse
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		if out.LatencyMS >= 100 {
			t.Fatalf("single inference latency too high: %dms", out.LatencyMS)
		}
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	p95 := lat[int(float64(len(lat))*0.95)]
	if p95 > 100*time.Millisecond {
		t.Fatalf("p95 latency %s exceeds 100ms", p95)
	}
}
