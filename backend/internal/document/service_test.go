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

	"bitbucket.org/HyperEther/document-management-system/internal/platform/completer"
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

func newSvc(repo *mockRepository, store *mockStorage) Service {
	return NewService(repo, store, queue.NewMemoryQueue(100), &mockEmbedder{}, &mockCompleter{})
}

func TestService_Create_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockRepository{
		createFn: func(ctx context.Context, doc *Document) error { return nil },
	}
	store := &mockStorage{
		uploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string) error { return nil },
	}

	header, file := buildFileHeader(t, "test.pdf", "application/pdf", "content")
	resp, err := newSvc(repo, store).Create(context.Background(), userID, header, file)

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
	header, file := buildFileHeader(t, "virus.exe", "application/octet-stream", "binary")
	_, err := newSvc(&mockRepository{}, &mockStorage{}).Create(context.Background(), uuid.New(), header, file)

	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType, got %v", err)
	}
}

func TestService_Create_MIMEMismatch(t *testing.T) {
	header, file := buildFileHeader(t, "sneaky.pdf", "text/plain", "not a pdf")
	_, err := newSvc(&mockRepository{}, &mockStorage{}).Create(context.Background(), uuid.New(), header, file)

	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("expected ErrUnsupportedType for MIME mismatch, got %v", err)
	}
}

func TestService_Create_StorageFailure_NoOrphan(t *testing.T) {
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
	_, err := newSvc(repo, store).Create(context.Background(), uuid.New(), header, file)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if repoCalled {
		t.Error("repo.Create should not be called when upload fails")
	}
}

func TestService_Create_DBFailure_CleansUpStorage(t *testing.T) {
	deleteKey := ""
	store := &mockStorage{
		uploadFn: func(ctx context.Context, key string, reader io.Reader, contentType string) error { return nil },
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
	_, err := newSvc(repo, store).Create(context.Background(), uuid.New(), header, file)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if deleteKey == "" {
		t.Error("expected storage.Delete to be called for cleanup, but it wasn't")
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
		countFn: func(ctx context.Context, uid uuid.UUID) (int, error) { return 2, nil },
	}

	result, err := newSvc(repo, &mockStorage{}).List(context.Background(), userID, pagination.New(1, 15))

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

func TestService_Get_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, Name: "test.pdf"}, nil
		},
	}

	resp, err := newSvc(repo, &mockStorage{}).Get(context.Background(), userID, docID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.ID != docID {
		t.Errorf("expected %s, got %s", docID, resp.ID)
	}
}

func TestService_Get_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	_, err := newSvc(repo, &mockStorage{}).Get(context.Background(), uuid.New(), uuid.New())

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

	_, err := newSvc(repo, &mockStorage{}).Get(context.Background(), uuid.New(), uuid.New())

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
		updateFn: func(ctx context.Context, doc *Document) error { return nil },
	}

	resp, err := newSvc(repo, &mockStorage{}).Update(context.Background(), userID, docID, &UpdateRequest{Name: "new.pdf"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Name != "new.pdf" {
		t.Errorf("expected new.pdf, got %s", resp.Name)
	}
}

func TestService_Update_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	_, err := newSvc(repo, &mockStorage{}).Update(context.Background(), uuid.New(), uuid.New(), &UpdateRequest{Name: "hacked.pdf"})

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Delete ---

func TestService_Delete_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, StorageKey: "key.pdf"}, nil
		},
		deleteFn: func(ctx context.Context, id uuid.UUID) error { return nil },
	}
	store := &mockStorage{
		deleteFn: func(ctx context.Context, key string) error { return nil },
	}

	if err := newSvc(repo, store).Delete(context.Background(), userID, docID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Delete_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	err := newSvc(repo, &mockStorage{}).Delete(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Download ---

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

	doc, reader, err := newSvc(repo, store).Download(context.Background(), userID, docID)
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
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID}, nil
		},
	}

	_, _, err := newSvc(repo, &mockStorage{}).Download(context.Background(), uuid.New(), uuid.New())

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Validation ---

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

func readyDoc(userID, docID uuid.UUID) *mockRepository {
	return &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, Status: StatusReady}, nil
		},
	}
}

func TestService_Query_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	repo := readyDoc(userID, docID)
	repo.findSimilarChunksFn = func(ctx context.Context, id uuid.UUID, embedding []float32, limit int) ([]*completer.SimilarChunk, error) {
		return []*completer.SimilarChunk{
			{Content: "relevant content"},
		}, nil
	}

	completer := &mockCompleter{
		completeFn: func(ctx context.Context, question string, chunks []*completer.SimilarChunk) (string, error) {
			return "the answer", nil
		},
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, completer)
	resp, err := svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "what is this?"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Answer != "the answer" {
		t.Errorf("expected 'the answer', got %s", resp.Answer)
	}
	if resp.Question != "what is this?" {
		t.Errorf("expected question echoed back, got %s", resp.Question)
	}
}

func TestService_Query_DocumentNotReady(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: docID, UserID: userID, Status: StatusProcessing}, nil
		},
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, &mockCompleter{})
	_, err := svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "?"})

	if !errors.Is(err, ErrNotReady) {
		t.Errorf("expected ErrNotReady, got %v", err)
	}
}

func TestService_Query_Forbidden(t *testing.T) {
	ownerID := uuid.New()
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return &Document{ID: uuid.New(), UserID: ownerID, Status: StatusReady}, nil
		},
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, &mockCompleter{})
	_, err := svc.Query(context.Background(), uuid.New(), uuid.New(), &QueryRequest{Question: "?"})

	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Query_NotFound(t *testing.T) {
	repo := &mockRepository{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*Document, error) {
			return nil, ErrNotFound
		},
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, &mockCompleter{})
	_, err := svc.Query(context.Background(), uuid.New(), uuid.New(), &QueryRequest{Question: "?"})

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Query_EmbedFailure(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, chunks []string) ([][]float32, error) {
			return nil, errors.New("openai unavailable")
		},
	}

	svc := NewService(readyDoc(userID, docID), &mockStorage{}, queue.NewMemoryQueue(1), embedder, &mockCompleter{})
	_, err := svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "?"})

	if err == nil {
		t.Error("expected error on embed failure, got nil")
	}
}

func TestService_Query_NoChunksFound_ReturnsDefault(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	repo := readyDoc(userID, docID)
	repo.findSimilarChunksFn = func(ctx context.Context, id uuid.UUID, embedding []float32, limit int) ([]*completer.SimilarChunk, error) {
		return []*completer.SimilarChunk{}, nil
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, &mockCompleter{})
	resp, err := svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "?"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Answer == "" {
		t.Error("expected a default answer when no chunks found")
	}
}

func TestService_Query_CompleterFailure(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()

	repo := readyDoc(userID, docID)
	repo.findSimilarChunksFn = func(ctx context.Context, id uuid.UUID, embedding []float32, limit int) ([]*completer.SimilarChunk, error) {
		return []*completer.SimilarChunk{{Content: "some content"}}, nil
	}

	completer := &mockCompleter{
		completeFn: func(ctx context.Context, question string, chunks []*completer.SimilarChunk) (string, error) {
			return "", errors.New("openai unavailable")
		},
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), &mockEmbedder{}, completer)
	_, err := svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "?"})

	if err == nil {
		t.Error("expected error on completer failure, got nil")
	}
}

func TestService_Query_EmbeddingPassedToSearch(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	var capturedEmbedding []float32

	fakeEmbedding := []float32{0.1, 0.2, 0.3}
	embedder := &mockEmbedder{
		embedFn: func(ctx context.Context, chunks []string) ([][]float32, error) {
			return [][]float32{fakeEmbedding}, nil
		},
	}

	repo := readyDoc(userID, docID)
	repo.findSimilarChunksFn = func(ctx context.Context, id uuid.UUID, embedding []float32, limit int) ([]*completer.SimilarChunk, error) {
		capturedEmbedding = embedding
		return []*completer.SimilarChunk{}, nil
	}

	svc := NewService(repo, &mockStorage{}, queue.NewMemoryQueue(1), embedder, &mockCompleter{})
	svc.Query(context.Background(), userID, docID, &QueryRequest{Question: "test question"})

	if len(capturedEmbedding) != len(fakeEmbedding) {
		t.Fatalf("expected embedding length %d, got %d", len(fakeEmbedding), len(capturedEmbedding))
	}
	for i, v := range fakeEmbedding {
		if capturedEmbedding[i] != v {
			t.Errorf("embedding[%d]: expected %f, got %f", i, v, capturedEmbedding[i])
		}
	}
}
