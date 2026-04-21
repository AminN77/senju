package queue

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func TestNATSQueue_RetryAndDeadLetter_NoSilentLoss(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}

	srv := natsserver.New(&natsserver.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
	})
	srv.Start()
	t.Cleanup(srv.Shutdown)
	if !srv.ReadyForConnections(10 * time.Second) {
		t.Fatal("nats server not ready")
	}

	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)

	cfg := Config{
		StreamName:   "jobs_stream",
		Subject:      "jobs.execute",
		ConsumerName: "jobs_worker",
		DeadLetter:   "jobs.dead_letter",
		MaxRetries:   2,
		BackoffBase:  20 * time.Millisecond,
	}
	q, err := NewNATSQueue(nc, cfg)
	if err != nil {
		t.Fatal(err)
	}

	js, err := nc.JetStream()
	if err != nil {
		t.Fatal(err)
	}

	dlqCh := make(chan map[string]any, 1)
	dlqSub, err := js.Subscribe(cfg.DeadLetter, func(msg *nats.Msg) {
		var payload map[string]any
		if err := json.Unmarshal(msg.Data, &payload); err == nil {
			select {
			case dlqCh <- payload:
			default:
			}
		}
		_ = msg.Ack()
	}, nats.ManualAck(), nats.Durable("dlq_observer"), nats.DeliverNew())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dlqSub.Unsubscribe() })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var attemptCounter int32
	errCh := make(chan error, 1)
	go func() {
		errCh <- q.Consume(ctx, func(context.Context, Message) error {
			atomic.AddInt32(&attemptCounter, 1)
			return context.DeadlineExceeded // failure injection
		})
	}()

	if err := q.Enqueue(context.Background(), Message{
		JobID:   "job-9",
		Payload: json.RawMessage(`{"stage":"qc"}`),
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case dlq := <-dlqCh:
		if dlq["job_id"] != "job-9" {
			t.Fatalf("unexpected dlq payload: %+v", dlq)
		}
		if dlq["dead_letter"] != true {
			t.Fatalf("expected dead_letter=true got %+v", dlq)
		}
		attempts, ok := dlq["attempts"].(float64)
		if !ok || int(attempts) != 3 {
			t.Fatalf("expected attempts=3 got %+v", dlq["attempts"])
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for dead-letter message")
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("consume returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("consume did not stop after cancellation")
	}

	if got := atomic.LoadInt32(&attemptCounter); got != 3 {
		t.Fatalf("expected 3 delivery attempts, got %d", got)
	}
}
