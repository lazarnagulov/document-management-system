package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMemoryQueue_PushAndReceive(t *testing.T) {
	q := NewMemoryQueue(10)
	job := Job{DocumentID: uuid.New()}

	if err := q.Push(context.Background(), job); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	select {
	case received := <-q.Jobs():
		if received.DocumentID != job.DocumentID {
			t.Errorf("expected %s, got %s", job.DocumentID, received.DocumentID)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for job")
	}
}

func TestMemoryQueue_Full(t *testing.T) {
	q := NewMemoryQueue(1)
	job := Job{DocumentID: uuid.New()}

	if err := q.Push(context.Background(), job); err != nil {
		t.Fatalf("first push failed: %v", err)
	}

	if err := q.Push(context.Background(), job); err == nil {
		t.Error("expected error on full queue, got nil")
	}
}

func TestMemoryQueue_CancelledContext(t *testing.T) {
	q := NewMemoryQueue(0) // zero buffer will always block

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := q.Push(ctx, Job{DocumentID: uuid.New()})
	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}
