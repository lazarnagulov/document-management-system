package processing

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var testLogger = zap.NewNop()

func TestProcessor_Process_Success(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	repo := successRepo(docID, &statuses)
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("hello world")), nil
		},
	}
	factory := extractor.NewFactory(extractor.Config{
		PDFURL:  "http://localhost",
		DOCXURL: "http://localhost",
	})

	p := NewProcessor(repo, store, factory, testLogger)
	p.Process(context.Background(), docID)

	if len(statuses) != 2 {
		t.Fatalf("expected 2 status updates, got %d", len(statuses))
	}
	if statuses[0] != document.StatusProcessing {
		t.Errorf("expected first status PROCESSING, got %s", statuses[0])
	}
	if statuses[1] != document.StatusReady {
		t.Errorf("expected second status READY, got %s", statuses[1])
	}
}

func TestProcessor_Process_DownloadFailure_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	repo := successRepo(docID, &statuses)
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return nil, errors.New("minio unavailable")
		},
	}
	factory := extractor.NewFactory(extractor.Config{
		PDFURL:  "http://localhost",
		DOCXURL: "http://localhost",
	})

	p := NewProcessor(repo, store, factory, testLogger)
	p.Process(context.Background(), docID)

	last := statuses[len(statuses)-1]
	if last != document.StatusFailed {
		t.Errorf("expected final status FAILED, got %s", last)
	}
}

func TestProcessor_Process_DocumentNotFound_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return nil, document.ErrNotFound
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			statuses = append(statuses, status)
			return nil
		},
	}

	p := NewProcessor(repo, &mockStorage{}, extractor.NewFactory(extractor.Config{}), testLogger)
	p.Process(context.Background(), docID)

	last := statuses[len(statuses)-1]
	if last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_Process_StatusTransitionOrder(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	repo := successRepo(docID, &statuses)
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("text content")), nil
		},
	}

	p := NewProcessor(repo, store, extractor.NewFactory(extractor.Config{
		PDFURL:  "http://localhost",
		DOCXURL: "http://localhost",
	}), testLogger)
	p.Process(context.Background(), docID)

	if len(statuses) < 1 || statuses[0] != document.StatusProcessing {
		t.Errorf("expected PROCESSING as first status transition, got %v", statuses)
	}
}
