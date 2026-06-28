package processing

import (
	"context"
	"fmt"

	"bitbucket.org/HyperEther/document-management-system/internal/document"
	"bitbucket.org/HyperEther/document-management-system/internal/extractor"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Processor struct {
	repo      document.Repository
	storage   storage.Storage
	extractor *extractor.Factory
	log       *zap.Logger
}

func NewProcessor(repo document.Repository, storage storage.Storage, extractor *extractor.Factory, log *zap.Logger) *Processor {
	return &Processor{
		repo:      repo,
		storage:   storage,
		extractor: extractor,
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

	doc, err := p.repo.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	file, err := p.storage.Download(ctx, doc.StorageKey)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	defer file.Close()

	ext, err := p.extractor.For(doc.ContentType)
	if err != nil {
		return fmt.Errorf("get extractor: %w", err)
	}

	text, err := ext.Extract(ctx, file)
	if err != nil {
		return fmt.Errorf("extract text: %w", err)
	}

	_ = text

	if err := p.repo.UpdateStatus(ctx, documentID, document.StatusReady); err != nil {
		return fmt.Errorf("set ready status: %w", err)
	}

	p.log.Info("document processed successfully", zap.String("document_id", documentID.String()))
	return nil
}
