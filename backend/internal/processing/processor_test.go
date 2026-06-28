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

func textDoc(docID uuid.UUID, contentType string) *document.Document {
	return &document.Document{
		ID:          docID,
		StorageKey:  "users/x/documents/y." + contentType,
		ContentType: contentType,
	}
}

func trackingRepo(docID uuid.UUID, statuses *[]document.DocumentStatus) *mockRepo {
	return &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return textDoc(docID, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			*statuses = append(*statuses, status)
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			return nil
		},
	}
}

func contentStore(text string) *mockStorage {
	return &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(text)), nil
		},
	}
}

func TestProcessor_Success_StatusTransitions(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	p := NewProcessor(
		trackingRepo(docID, &statuses),
		contentStore("hello world"),
		&mockExtractorFactory{},
		&mockChunker{},
		&mockEmbedder{},
		testLogger,
	)
	p.Process(context.Background(), docID)

	if len(statuses) != 2 {
		t.Fatalf("expected 2 status updates, got %d: %v", len(statuses), statuses)
	}
	if statuses[0] != document.StatusProcessing {
		t.Errorf("expected PROCESSING first, got %s", statuses[0])
	}
	if statuses[1] != document.StatusReady {
		t.Errorf("expected READY second, got %s", statuses[1])
	}
}

func TestProcessor_ProcessingIsFirstStatus(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	p := NewProcessor(
		trackingRepo(docID, &statuses),
		contentStore("some text"),
		&mockExtractorFactory{},
		&mockChunker{},
		&mockEmbedder{},
		testLogger,
	)
	p.Process(context.Background(), docID)

	if len(statuses) == 0 || statuses[0] != document.StatusProcessing {
		t.Errorf("expected PROCESSING as first status, got %v", statuses)
	}
}

func TestProcessor_DocumentNotFound_SetsFailed(t *testing.T) {
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

	p := NewProcessor(repo, &mockStorage{}, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_DownloadFailure_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return nil, errors.New("minio unavailable")
		},
	}

	p := NewProcessor(trackingRepo(docID, &statuses), store, &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_ExtractionFailure_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	factory := &mockExtractorFactory{
		forFn: func(contentType string) (extractor.Extractor, error) {
			return &mockExtractor{
				extractFn: func(ctx context.Context, content io.Reader) (string, error) {
					return "", errors.New("extraction failed")
				},
			}, nil
		},
	}

	p := NewProcessor(
		trackingRepo(docID, &statuses),
		contentStore("some content"),
		factory,
		&mockChunker{},
		&mockEmbedder{},
		testLogger,
	)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_UnsupportedContentType_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	factory := &mockExtractorFactory{
		forFn: func(contentType string) (extractor.Extractor, error) {
			return nil, errors.New("unsupported content type")
		},
	}

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return &document.Document{
				ID:          id,
				StorageKey:  "key",
				ContentType: "application/unknown",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			statuses = append(statuses, status)
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error { return nil },
	}

	p := NewProcessor(repo, contentStore("data"), factory, &mockChunker{}, &mockEmbedder{}, testLogger)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_EmbedFailure_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, chunks []string) ([][]float32, error) {
			return nil, errors.New("openai unavailable")
		},
	}

	p := NewProcessor(
		trackingRepo(docID, &statuses),
		contentStore("some text to embed"),
		&mockExtractorFactory{},
		&mockChunker{},
		embedder,
		testLogger,
	)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_SaveChunksFailure_SetsFailed(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return textDoc(docID, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			statuses = append(statuses, status)
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			return errors.New("db write failed")
		},
	}

	p := NewProcessor(repo, contentStore("some text"), &mockExtractorFactory{}, &mockChunker{}, &mockEmbedder{}, testLogger)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusFailed {
		t.Errorf("expected FAILED, got %s", last)
	}
}

func TestProcessor_EmptyTextAfterExtraction_StillReady(t *testing.T) {
	docID := uuid.New()
	var statuses []document.DocumentStatus

	chunker := &mockChunker{
		chunkFn: func(text string) []string { return nil },
	}

	p := NewProcessor(
		trackingRepo(docID, &statuses),
		contentStore(""),
		&mockExtractorFactory{},
		chunker,
		&mockEmbedder{},
		testLogger,
	)
	p.Process(context.Background(), docID)

	if last := statuses[len(statuses)-1]; last != document.StatusReady {
		t.Errorf("expected READY for blank document, got %s", last)
	}
}

func TestProcessor_ChunksStoredWithCorrectDocumentID(t *testing.T) {
	docID := uuid.New()
	var savedChunks []*document.Chunk

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return textDoc(docID, "text/plain"), nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			return nil
		},
		saveChunksFn: func(ctx context.Context, chunks []*document.Chunk) error {
			savedChunks = chunks
			return nil
		},
	}

	chunker := &mockChunker{
		chunkFn: func(text string) []string {
			return []string{"chunk one", "chunk two", "chunk three"}
		},
	}

	p := NewProcessor(repo, contentStore("some text"), &mockExtractorFactory{}, chunker, &mockEmbedder{}, testLogger)
	p.Process(context.Background(), docID)

	if len(savedChunks) != 3 {
		t.Fatalf("expected 3 chunks saved, got %d", len(savedChunks))
	}
	for i, chunk := range savedChunks {
		if chunk.DocumentID != docID {
			t.Errorf("chunk %d has wrong document ID: %s", i, chunk.DocumentID)
		}
		if chunk.ChunkIndex != i {
			t.Errorf("chunk %d has wrong index: %d", i, chunk.ChunkIndex)
		}
		if chunk.ID == (uuid.UUID{}) {
			t.Errorf("chunk %d has zero UUID", i)
		}
	}
}

func TestProcessor_EmbeddingCountMatchesChunkCount(t *testing.T) {
	docID := uuid.New()
	var embedCallChunks []string

	chunker := &mockChunker{
		chunkFn: func(text string) []string {
			return []string{"a", "b", "c"}
		},
	}
	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, chunks []string) ([][]float32, error) {
			embedCallChunks = chunks
			result := make([][]float32, len(chunks))
			for i := range result {
				result[i] = make([]float32, 1536)
			}
			return result, nil
		},
	}

	p := NewProcessor(
		trackingRepo(docID, &[]document.DocumentStatus{}),
		contentStore("text"),
		&mockExtractorFactory{},
		chunker,
		embedder,
		testLogger,
	)
	p.Process(context.Background(), docID)

	if len(embedCallChunks) != 3 {
		t.Errorf("expected 3 chunks sent to embedder, got %d", len(embedCallChunks))
	}
}
