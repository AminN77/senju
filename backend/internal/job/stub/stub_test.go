package stub

import (
	"context"
	"testing"

	"github.com/AminN77/senju/backend/internal/job"
)

func TestStub_CreateGetUpdate(t *testing.T) {
	t.Parallel()
	r := New()
	ctx := context.Background()

	created, err := r.Create(ctx, job.CreateParams{
		Status: job.StatusPending,
		Stage:  "qc",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != job.StatusPending {
		t.Fatalf("status %q", got.Status)
	}
	updated, err := r.UpdateStatusStage(ctx, created.ID, job.StatusRunning, "align")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != job.StatusRunning || updated.Stage != "align" {
		t.Fatalf("got %+v", updated)
	}
}
