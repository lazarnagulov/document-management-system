package processing

import (
	"context"
	"fmt"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/chunker"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/embedder"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Processor struct {
	repo      document.Repository
	storage   storage.Storage
	extractor extractor.Factory
	chunker   chunker.Chunker
	embedder  embedder.Embedder
	log       *zap.Logger
}

func NewProcessor(
	repo document.Repository,
	storage storage.Storage,
	extractor extractor.Factory,
	chunker chunker.Chunker,
	embedder embedder.Embedder,
	log *zap.Logger,
) *Processor {
	return &Processor{
		repo:      repo,
		storage:   storage,
		extractor: extractor,
		embedder:  embedder,
		chunker:   chunker,
		log:       log,
	}
}

func (p *Processor) Process(ctx context.Context, documentID uuid.UUID) {
	if err := p.process(ctx, documentID); err != nil {
		p.log.Error("processing failed",
			zap.String("document_id", documentID.String()),
			zap.Error(err),
		)

		if err := p.repo.UpdateStatus(ctx, documentID, document.StatusFailed); err != nil {
			p.log.Error("failed to set document status to FAILED",
				zap.String("document_id", documentID.String()),
				zap.Error(err),
			)
		}
	}
}

func (p *Processor) process(ctx context.Context, documentID uuid.UUID) error {
	p.log.Info("starting document processing", zap.String("document_id", documentID.String()))

	if err := p.repo.UpdateStatus(ctx, documentID, document.StatusProcessing); err != nil {
		return fmt.Errorf("set processing status: %w", err)
	}

	doc, text, err := p.loadDocumentText(ctx, documentID)
	if err != nil {
		return err
	}

	rawChunks := p.chunker.Chunk(text)
	if len(rawChunks) == 0 {
		return p.repo.UpdateStatus(ctx, documentID, document.StatusReady)
	}

	chunks, err := p.createChunks(ctx, doc.ID, rawChunks)
	if err != nil {
		return err
	}

	if err := p.repo.SaveChunks(ctx, chunks); err != nil {
		return fmt.Errorf("save chunks: %w", err)
	}

	return p.repo.UpdateStatus(ctx, documentID, document.StatusReady)
}

func (p *Processor) loadDocumentText(ctx context.Context, documentID uuid.UUID) (*document.Document, string, error) {
	doc, err := p.repo.GetByID(ctx, documentID)
	if err != nil {
		return nil, "", fmt.Errorf("get document: %w", err)
	}

	file, err := p.storage.Download(ctx, doc.StorageKey)
	if err != nil {
		return nil, "", fmt.Errorf("download file: %w", err)
	}
	defer file.Close()

	ext, err := p.extractor.For(doc.ContentType)
	if err != nil {
		return nil, "", fmt.Errorf("get extractor: %w", err)
	}

	text, err := ext.Extract(ctx, file)
	if err != nil {
		return nil, "", fmt.Errorf("extract text: %w", err)
	}

	return doc, text, nil
}

func (p *Processor) createChunks(ctx context.Context, documentID uuid.UUID, rawChunks []string) ([]*document.Chunk, error) {
	embeddings, err := p.embedder.Embed(ctx, rawChunks)
	if err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}

	now := time.Now()
	chunks := make([]*document.Chunk, len(rawChunks))

	for i, content := range rawChunks {
		chunks[i] = &document.Chunk{
			ID:         uuid.New(),
			DocumentID: documentID,
			Content:    content,
			Embedding:  embeddings[i],
			ChunkIndex: i,
			CreatedAt:  now,
		}
	}

	return chunks, nil
}
