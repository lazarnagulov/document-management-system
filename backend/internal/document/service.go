package document

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"bitbucket.org/HyperEther/document-management-system/internal/storage"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error)
	Get(ctx context.Context, userID, docID uuid.UUID) (*Response, error)
	List(ctx context.Context, userID uuid.UUID, page pagination.Page) (*pagination.Response[*Response], error)
	Update(ctx context.Context, userID, docID uuid.UUID, req *UpdateRequest) (*Response, error)
	Delete(ctx context.Context, userID, docID uuid.UUID) error
	Download(ctx context.Context, userID, docID uuid.UUID) (*Document, io.ReadCloser, error)
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{repo: repo, storage: storage}
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error) {
	info, err := validateFile(header)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        info.name,
		ContentType: info.contentType,
		Size:        info.size,
		StorageKey:  storageKey(userID, uuid.New(), info.ext),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.storage.Upload(ctx, doc.StorageKey, file, doc.ContentType); err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		_ = s.storage.Delete(ctx, doc.StorageKey)
		return nil, fmt.Errorf("save document: %w", err)
	}

	return ToResponse(doc), nil
}

func (s *service) Get(ctx context.Context, userID, docID uuid.UUID) (*Response, error) {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, ErrForbidden
	}
	return ToResponse(doc), nil
}

func (s *service) Update(ctx context.Context, userID, docID uuid.UUID, req *UpdateRequest) (*Response, error) {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.UserID != userID {
		return nil, ErrForbidden
	}

	doc.Name = req.Name

	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("update document: %w", err)
	}
	return ToResponse(doc), nil
}

func (s *service) Delete(ctx context.Context, userID, docID uuid.UUID) error {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Delete(ctx, docID); err != nil {
		return fmt.Errorf("delete document record: %w", err)
	}

	if err := s.storage.Delete(ctx, doc.StorageKey); err != nil {
		return fmt.Errorf("delete file (orphaned): %w", err)
	}

	return nil
}

func (s *service) Download(ctx context.Context, userID, docID uuid.UUID) (*Document, io.ReadCloser, error) {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return nil, nil, err
	}
	if doc.UserID != userID {
		return nil, nil, ErrForbidden
	}

	reader, err := s.storage.Download(ctx, doc.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("download file: %w", err)
	}

	return doc, reader, nil
}

func (s *service) List(ctx context.Context, userID uuid.UUID, page pagination.Page) (*pagination.Response[*Response], error) {
	docs, err := s.repo.ListByUserID(ctx, userID, page)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}

	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count documents: %w", err)
	}

	resp := pagination.NewResponse(ToResponseList(docs), total, page)
	return &resp, nil
}

func storageKey(userID, fileID uuid.UUID, ext string) string {
	return fmt.Sprintf("users/%s/documents/%s%s", userID, fileID, ext)
}
