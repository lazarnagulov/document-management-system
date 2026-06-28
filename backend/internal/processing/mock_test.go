package processing

import (
	"context"
	"errors"
	"fmt"
	"io"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/chunker"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"github.com/google/uuid"
)

type mockRepo struct {
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*document.Document, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error
	saveChunksFn   func(ctx context.Context, chunks []*document.Chunk) error
}

func (m *mockRepo) GetByID(ctx context.Context, id uuid.UUID) (*document.Document, error) {
	if m.getByIDFn == nil {
		return nil, errors.New("getByIDFn not set")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
	if m.updateStatusFn == nil {
		return errors.New("updateStatusFn not set")
	}
	return m.updateStatusFn(ctx, id, status)
}

func (m *mockRepo) SaveChunks(ctx context.Context, chunks []*document.Chunk) error {
	if m.saveChunksFn == nil {
		return fmt.Errorf("saveChunksFn not set")
	}
	return m.saveChunksFn(ctx, chunks)
}

func (m *mockRepo) Create(ctx context.Context, doc *document.Document) error { return nil }
func (m *mockRepo) Update(ctx context.Context, doc *document.Document) error { return nil }
func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error           { return nil }
func (m *mockRepo) ListByUserID(ctx context.Context, userID uuid.UUID, page pagination.Page) ([]*document.Document, error) {
	return nil, nil
}
func (m *mockRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) { return 0, nil }

type mockStorage struct {
	downloadFn func(ctx context.Context, key string) (io.ReadCloser, error)
}

func (m *mockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadFn == nil {
		return nil, errors.New("downloadFn not set")
	}
	return m.downloadFn(ctx, key)
}

func (m *mockStorage) Upload(ctx context.Context, key string, r io.Reader, ct string) error {
	return nil
}
func (m *mockStorage) Delete(ctx context.Context, key string) error { return nil }

func successRepo(docID uuid.UUID, statuses *[]document.DocumentStatus) *mockRepo {
	return &mockRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*document.Document, error) {
			return &document.Document{
				ID:          docID,
				StorageKey:  "users/x/documents/y.pdf",
				ContentType: "text/plain",
			}, nil
		},
		updateStatusFn: func(ctx context.Context, id uuid.UUID, status document.DocumentStatus) error {
			*statuses = append(*statuses, status)
			return nil
		},
	}
}

type mockExtractorFactory struct {
	forFn func(contentType string) (extractor.Extractor, error)
}

func (m *mockExtractorFactory) For(contentType string) (extractor.Extractor, error) {
	if m.forFn == nil {
		return &mockExtractor{}, nil
	}
	return m.forFn(contentType)
}

type mockExtractor struct {
	extractFn func(ctx context.Context, content io.Reader) (string, error)
}

func (m *mockExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	if m.extractFn == nil {
		raw, err := io.ReadAll(content)
		return string(raw), err
	}
	return m.extractFn(ctx, content)
}

type mockChunker struct {
	chunkFn func(text string) []string
}

func (m *mockChunker) Chunk(text string) []string {
	if m.chunkFn == nil {
		return []string{text}
	}
	return m.chunkFn(text)
}

func (c *mockChunker) ChunkSize() int {
	return chunker.DefaultChunkSize
}

func (c *mockChunker) ChunkOverlap() int {
	return chunker.DefaultChunkOverlap
}

type mockEmbedder struct {
	embedFn func(ctx context.Context, chunks []string) ([][]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, chunks []string) ([][]float32, error) {
	if m.embedFn == nil {
		result := make([][]float32, len(chunks))
		for i := range result {
			result[i] = make([]float32, 1536)
		}
		return result, nil
	}
	return m.embedFn(ctx, chunks)
}
