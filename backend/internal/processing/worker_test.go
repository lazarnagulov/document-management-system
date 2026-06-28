package processing

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/queue"
	"github.com/google/uuid"
)

func TestWorker_ProcessesJob(t *testing.T) {
	docID := uuid.New()
	ready := make(chan uuid.UUID, 1)

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return textDoc(id, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			if status == document.StatusReady {
				ready <- id
			}
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			return nil
		},
	}
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}

	q := queue.NewMemoryQueue(10)
	processor := NewProcessor(repo, store, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	NewWorker(q, processor).Start(ctx)
	q.Push(ctx, queue.Job{DocumentID: docID})

	select {
	case id := <-ready:
		if id != docID {
			t.Errorf("expected %s, got %s", docID, id)
		}
	case <-time.After(3 * time.Second):
		t.Error("timed out waiting for job to be processed")
	}
}

func TestWorker_ProcessesMultipleJobs(t *testing.T) {
	processed := make(chan uuid.UUID, 3)
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return textDoc(id, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			if status == document.StatusReady {
				processed <- id
			}
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			return nil
		},
	}
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}

	q := queue.NewMemoryQueue(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	NewWorker(q, NewProcessor(repo, store, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)).Start(ctx)

	for _, id := range ids {
		q.Push(ctx, queue.Job{DocumentID: id})
	}

	seen := make(map[uuid.UUID]bool)
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(ids); i++ {
		select {
		case id := <-processed:
			seen[id] = true
		case <-timeout:
			t.Fatalf("timed out: only processed %d of %d jobs", len(seen), len(ids))
		}
	}

	for _, id := range ids {
		if !seen[id] {
			t.Errorf("job %s was never processed", id)
		}
	}
}

func TestWorker_StopsOnContextCancel(t *testing.T) {
	q := queue.NewMemoryQueue(10)
	processor := NewProcessor(&mockRepo{}, &mockStorage{}, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)

	ctx, cancel := context.WithCancel(context.Background())
	NewWorker(q, processor).Start(ctx)
	cancel()

	time.Sleep(100 * time.Millisecond)

	q.Push(context.Background(), queue.Job{DocumentID: uuid.New()})
}

func TestWorker_FailedJobDoesNotBlockQueue(t *testing.T) {
	processed := make(chan uuid.UUID, 2)
	failID := uuid.New()
	successID := uuid.New()

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			if id == failID {
				return nil, document.ErrNotFound
			}
			return textDoc(id, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			if status == document.StatusReady || status == document.StatusFailed {
				processed <- id
			}
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			return nil
		},
	}
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}

	q := queue.NewMemoryQueue(10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	NewWorker(q, NewProcessor(repo, store, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)).Start(ctx)

	q.Push(ctx, queue.Job{DocumentID: failID})
	q.Push(ctx, queue.Job{DocumentID: successID})

	seen := make(map[uuid.UUID]bool)
	timeout := time.After(5 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case id := <-processed:
			seen[id] = true
		case <-timeout:
			t.Fatalf("timed out waiting for jobs, seen: %v", seen)
		}
	}

	if !seen[successID] {
		t.Error("success job was blocked by failed job")
	}
}
