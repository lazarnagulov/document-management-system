package document

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/queue"
	"github.com/google/uuid"
)

func buildFileHeader(t *testing.T, filename, contentType, content string) (*multipart.FileHeader, io.Reader) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="document"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("could not create part: %v", err)
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		t.Fatalf("could not write content: %v", err)
	}
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		t.Fatalf("could not read form: %v", err)
	}

	headers := form.File["document"]
	if len(headers) == 0 {
		t.Fatal("no file in form")
	}

	f, err := headers[0].Open()
	if err != nil {
		t.Fatalf("could not open file header: %v", err)
	}

	return headers[0], f
}

func TestService_Create_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		createFn: func(ctx context.Context, doc *Document) error {
			return nil
		},
	}
	store := &mockStorage{
		uploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string) error {
			return nil
		},
	}

	header, file := buildFileHeader(t, "test.pdf", "application/pdf", "content")
	svc := NewService(repo, store, queue.NewMemoryQueue(100))
	resp, err := svc.Create(context.Background(), userID, header, file)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != "test.pdf" {
		t.Errorf("expected test.pdf, got %s", resp.Name)
	}
	if resp.ContentType != "application/pdf" {
		t.Errorf("expected application/pdf, got %s", resp.ContentType)
	}
	if resp.ID == (uuid.UUID{}) {
		t.Error("expected non-zero document ID")
	}
}

func TestService_Create_UnsupportedType(t *testing.T) {
	repo := &mockRepository{}
	store := &mockStorage{}

	header, file := buildFileHeader(t, "virus.exe", "application/octet-stream", "binary")
	_, err := NewService(repo, store, queue.NewMemoryQueue(1000)).Create(context.Background(), uuid.New(), header, file)

	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestService_Create_MIMEMismatch(t *testing.T) {
	repo := &mockRepository{}
	store := &mockStorage{}

	header, file := buildFileHeader(t, "sneaky.pdf", "text/plain", "not a pdf")
	_, err := NewService(repo, store, queue.NewMemoryQueue(1000)).Create(context.Background(), uuid.New(), header, file)

	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType for MIME mismatch, got %v", err)
	}
}

func TestService_Create_StorageFailure_NoOrphan(t *testing.T) {
	// if upload fails, repo.Create should never be called
	repoCalled := false
	repo := &mockRepository{
		createFn: func(ctx context.Context, doc *Document) error {
			repoCalled = true
			return nil
		},
	}
	store := &mockStorage{
		uploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string) error {
			return errors.New("minio unavailable")
		},
	}

	header, file := buildFileHeader(t, "test.pdf", "application/pdf", "content")
	_, err := NewService(repo, store, queue.NewMemoryQueue(1000)).Create(context.Background(), uuid.New(), header, file)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if repoCalled {
		t.Error("repo.Create should not be called when upload fails")
	}
}

func TestService_Create_DBFailure_CleansUpStorage(t *testing.T) {
	// if DB insert fails after upload, storage.Delete should be called
	deleteKey := ""
	store := &mockStorage{
		uploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string) error {
			return nil
		},
		deleteFn: func(ctx context.Context, key string) error {
			deleteKey = key
			return nil
		},
	}
	repo := &mockRepository{
		createFn: func(ctx context.Context, doc *Document) error {
			return errors.New("db insert failed")
		},
	}

	header, file := buildFileHeader(t, "test.pdf", "application/pdf", "content")
	_, err := NewService(repo, store, queue.NewMemoryQueue(1000)).Create(context.Background(), uuid.New(), header, file)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if deleteKey == "" {
		t.Error("expected storage.Delete to be called for cleanup, but it wasn't")
	}
}

func TestService_Get_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, Name: "test.pdf"}, nil
		},
	}

	resp, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Get(context.Background(), userID, docID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ID != docID {
		t.Errorf("expected %s, got %s", docID, resp.ID)
	}
}

func TestService_Get_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	attackerID := uuid.New()
	docID := uuid.New()

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: ownerID}, nil
		},
	}

	_, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Get(context.Background(), attackerID, docID)

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return nil, ErrNotFound
		},
	}

	_, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Get(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, Name: "old.pdf"}, nil
		},
		updateFn: func(ctx context.Context, doc *Document) error {
			return nil
		},
	}

	resp, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Update(
		context.Background(), userID, docID, &UpdateRequest{Name: "new.pdf"},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != "new.pdf" {
		t.Errorf("expected new.pdf, got %s", resp.Name)
	}
}

func TestService_Update_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	attackerID := uuid.New()
	docID := uuid.New()

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: ownerID}, nil
		},
	}

	_, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Update(
		context.Background(), attackerID, docID, &UpdateRequest{Name: "hacked.pdf"},
	)

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Delete_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	storageKey := "users/x/documents/y.pdf"

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, StorageKey: storageKey}, nil
		},
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	store := &mockStorage{
		deleteFn: func(ctx context.Context, key string) error {
			return nil
		},
	}

	if err := NewService(repo, store, queue.NewMemoryQueue(1000)).Delete(context.Background(), userID, docID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	attackerID := uuid.New()

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Delete(context.Background(), attackerID, uuid.New())

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Download_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	content := "pdf content"

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, StorageKey: "key.pdf"}, nil
		},
	}
	store := &mockStorage{
		downloadFn: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(content)), nil
		},
	}

	doc, reader, err := NewService(repo, store, queue.NewMemoryQueue(1000)).Download(context.Background(), userID, docID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer reader.Close()

	if doc.ID != docID {
		t.Errorf("expected doc ID %s, got %s", docID, doc.ID)
	}

	got, _ := io.ReadAll(reader)
	if string(got) != content {
		t.Errorf("expected content %q, got %q", content, got)
	}
}

func TestService_Download_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	attackerID := uuid.New()

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	_, _, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).Download(context.Background(), attackerID, uuid.New())

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestValidateFile_AllowedTypes(t *testing.T) {
	cases := []struct {
		filename    string
		contentType string
	}{
		{"doc.pdf", "application/pdf"},
		{"doc.txt", "text/plain"},
		{"doc.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	}

	for _, tc := range cases {
		t.Run(tc.filename, func(t *testing.T) {
			header, _ := buildFileHeader(t, tc.filename, tc.contentType, "content")
			info, err := validateFile(header)
			if err != nil {
				t.Errorf("expected no error for %s, got %v", tc.filename, err)
			}
			if info.contentType != tc.contentType {
				t.Errorf("expected %s, got %s", tc.contentType, info.contentType)
			}
		})
	}
}

func TestValidateFile_RejectedTypes(t *testing.T) {
	cases := []struct {
		filename    string
		contentType string
	}{
		{"malware.exe", "application/octet-stream"},
		{"image.png", "image/png"},
		{"archive.zip", "application/zip"},
		{"page.html", "text/html"},
	}

	for _, tc := range cases {
		t.Run(tc.filename, func(t *testing.T) {
			header, _ := buildFileHeader(t, tc.filename, tc.contentType, "content")
			_, err := validateFile(header)
			if !errors.Is(err, ErrUnsupportedType) {
				t.Errorf("expected ErrUnsupportedType for %s, got %v", tc.filename, err)
			}
		})
	}
}

func TestService_List_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		listFn: func(ctx context.Context, uid uuid.UUID, page pagination.Page) ([]*Document, error) {
			return []*Document{
				{ID: uuid.New(), UserID: uid, Name: "a.pdf"},
				{ID: uuid.New(), UserID: uid, Name: "b.txt"},
			}, nil
		},
		countFn: func(ctx context.Context, uid uuid.UUID) (int, error) {
			return 2, nil
		},
	}

	page := pagination.New(1, 15)
	result, err := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1000)).List(context.Background(), userID, page)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Content) != 2 {
		t.Errorf("expected 2 docs, got %d", len(result.Content))
	}
	if result.TotalElements != 2 {
		t.Errorf("expected total 2, got %d", result.TotalElements)
	}
	if result.TotalPages != 1 {
		t.Errorf("expected 1 page, got %d", result.TotalPages)
	}
}
