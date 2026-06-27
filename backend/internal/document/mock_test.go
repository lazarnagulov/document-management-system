package document

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"github.com/google/uuid"
)

type mockService struct {
	createFn   func(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error)
	getFn      func(ctx context.Context, userID, docID uuid.UUID) (*Response, error)
	listFn     func(ctx context.Context, userID uuid.UUID, page pagination.Page) (*pagination.Response[*Response], error)
	updateFn   func(ctx context.Context, userID, docID uuid.UUID, req *UpdateRequest) (*Response, error)
	deleteFn   func(ctx context.Context, userID, docID uuid.UUID) error
	downloadFn func(ctx context.Context, userID, docID uuid.UUID) (*Document, io.ReadCloser, error)
}

func (m *mockService) Create(ctx context.Context, userID uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error) {
	if m.createFn == nil {
		return nil, fmt.Errorf("createFn not set")
	}
	return m.createFn(ctx, userID, header, file)
}

func (m *mockService) Get(ctx context.Context, userID, docID uuid.UUID) (*Response, error) {
	if m.getFn == nil {
		return nil, fmt.Errorf("getFn not set")
	}
	return m.getFn(ctx, userID, docID)
}

func (m *mockService) List(ctx context.Context, userID uuid.UUID, page pagination.Page) (*pagination.Response[*Response], error) {
	if m.listFn == nil {
		return nil, fmt.Errorf("listFn not set")
	}
	return m.listFn(ctx, userID, page)
}

func (m *mockService) Update(ctx context.Context, userID, docID uuid.UUID, req *UpdateRequest) (*Response, error) {
	if m.updateFn == nil {
		return nil, fmt.Errorf("updateFn not set")
	}
	return m.updateFn(ctx, userID, docID, req)
}

func (m *mockService) Delete(ctx context.Context, userID, docID uuid.UUID) error {
	if m.deleteFn == nil {
		return fmt.Errorf("deleteFn not set")
	}
	return m.deleteFn(ctx, userID, docID)
}

func (m *mockService) Download(ctx context.Context, userID, docID uuid.UUID) (*Document, io.ReadCloser, error) {
	if m.downloadFn == nil {
		return nil, nil, fmt.Errorf("downloadFn not set")
	}
	return m.downloadFn(ctx, userID, docID)
}

type mockRepository struct {
	createFn  func(ctx context.Context, doc *Document) error
	getByIDFn func(ctx context.Context, id uuid.UUID) (*Document, error)
	updateFn  func(ctx context.Context, doc *Document) error
	listFn    func(ctx context.Context, userID uuid.UUID, page pagination.Page) ([]*Document, error)
	countFn   func(ctx context.Context, userID uuid.UUID) (int, error)
	deleteFn  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRepository) Create(ctx context.Context, doc *Document) error {
	if m.createFn == nil {
		return fmt.Errorf("createFn not set")
	}
	return m.createFn(ctx, doc)
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	if m.getByIDFn == nil {
		return nil, fmt.Errorf("getByIDFn not set")
	}
	return m.getByIDFn(ctx, id)
}

func (m *mockRepository) Update(ctx context.Context, doc *Document) error {
	if m.updateFn == nil {
		return fmt.Errorf("updateFn not set")
	}
	return m.updateFn(ctx, doc)
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn == nil {
		return fmt.Errorf("deleteFn not set")
	}
	return m.deleteFn(ctx, id)
}

func (m *mockRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page pagination.Page) ([]*Document, error) {
	if m.listFn == nil {
		return nil, fmt.Errorf("listFn not set")
	}
	return m.listFn(ctx, userID, page)
}

func (m *mockRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.countFn == nil {
		return 0, fmt.Errorf("countFn not set")
	}
	return m.countFn(ctx, userID)
}

type mockStorage struct {
	uploadFn   func(ctx context.Context, key string, reader io.Reader, contentType string) error
	downloadFn func(ctx context.Context, key string) (io.ReadCloser, error)
	deleteFn   func(ctx context.Context, key string) error
}

func (m *mockStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	if m.uploadFn == nil {
		return fmt.Errorf("uploadFn not set")
	}
	return m.uploadFn(ctx, key, reader, contentType)
}

func (m *mockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadFn == nil {
		return nil, fmt.Errorf("downloadFn not set")
	}
	return m.downloadFn(ctx, key)
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	if m.deleteFn == nil {
		return fmt.Errorf("deleteFn not set")
	}
	return m.deleteFn(ctx, key)
}
