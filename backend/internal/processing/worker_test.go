package processing

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/queue"
	"github.com/google/uuid"
)

func TestWorker_ProcessesJob(t *testing.T) {
	docID := uuid.New()
	processed := make(chan uuid.UUID, 1)

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return &document.Document{
				ID:          id,
				StorageKey:  "key",
				ContentType: "text/plain",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			if status == document.StatusReady {
				processed <- id
			}
			return nil
		},
	}
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("content")), nil
		},
	}

	q := queue.NewMemoryQueue(10)
	processor := NewProcessor(repo, store, extractor.NewFactory(extractor.Config{
		PDFURL:  "http://localhost",
		DOCXURL: "http://localhost",
	}), testLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := NewWorker(q, processor)
	worker.Start(ctx)

	q.Push(ctx, queue.Job{DocumentID: docID})

	select {
	case id := <-processed:
		if id != docID {
			t.Errorf("expected %s, got %s", docID, id)
		}
	case <-time.After(3 * time.Second):
		t.Error("timed out waiting for job to be processed")
	}
}

func TestWorker_StopsOnContextCancel(t *testing.T) {
	q := queue.NewMemoryQueue(10)
	processor := NewProcessor(&mockRepo{}, &mockStorage{}, extractor.NewFactory(extractor.Config{}), testLogger)

	ctx, cancel := context.WithCancel(context.Background())
	worker := NewWorker(q, processor)
	worker.Start(ctx)
	cancel()

	time.Sleep(100 * time.Millisecond)
	q.Push(context.Background(), queue.Job{DocumentID: uuid.New()})
}
