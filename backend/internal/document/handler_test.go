package document

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(h *Handler) *gin.Engine {
	r := gin.New()
	rg := r.Group("/")
	h.RegisterRoutes(rg)
	return r
}

func injectUser(userID uuid.UUID, email string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.ContextUserKey, &token.Claims{
			UserID: userID,
			Email:  email,
		})
		c.Next()
	}
}

func newTestHandler(svc Service, userID uuid.UUID) *Handler {
	return NewHandler(svc, injectUser(userID, "test@example.com"))
}

func multipartFile(t *testing.T, fieldName, filename, contentType, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("could not create form file: %v", err)
	}
	if _, err := io.Copy(part, strings.NewReader(content)); err != nil {
		t.Fatalf("could not write file content: %v", err)
	}

	writer.Close()
	return body, writer.FormDataContentType()
}

func sampleDoc(userID uuid.UUID) *Response {
	return &Response{
		ID:          uuid.New(),
		Name:        "test.pdf",
		ContentType: "application/pdf",
		Size:        1024,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func TestHandler_Create_Success(t *testing.T) {
	userID := uuid.New()
	expected := sampleDoc(userID)

	svc := &mockService{
		createFn: func(ctx context.Context, uid uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error) {
			return expected, nil
		},
	}

	body, ct := multipartFile(t, "document", "test.pdf", "application/pdf", "%PDF-1.4 fake content")
	req := httptest.NewRequest(http.MethodPost, "/documents/", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Name != expected.Name {
		t.Errorf("expected name %s, got %s", expected.Name, resp.Name)
	}
}

func TestHandler_Create_MissingFile(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/documents/", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(&mockService{}, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_UnsupportedType(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		createFn: func(ctx context.Context, uid uuid.UUID, header *multipart.FileHeader, file io.Reader) (*Response, error) {
			return nil, ErrUnsupportedType
		},
	}

	body, ct := multipartFile(t, "document", "test.exe", "application/octet-stream", "binary content")
	req := httptest.NewRequest(http.MethodPost, "/documents/", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestHandler_Create_Unauthorized(t *testing.T) {
	// no middleware injected — simulates missing token
	h := NewHandler(&mockService{}, func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	body, ct := multipartFile(t, "document", "test.pdf", "application/pdf", "content")
	req := httptest.NewRequest(http.MethodPost, "/documents/", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	newTestRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandler_Get_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	svc := &mockService{
		getFn: func(ctx context.Context, uid, did uuid.UUID) (*Response, error) {
			return &Response{ID: docID, Name: "test.pdf"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String(), nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		getFn: func(ctx context.Context, uid, did uuid.UUID) (*Response, error) {
			return nil, ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/documents/"+uuid.NewString(), nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Get_Forbidden(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		getFn: func(ctx context.Context, uid, did uuid.UUID) (*Response, error) {
			return nil, ErrForbidden
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/documents/"+uuid.NewString(), nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/documents/not-a-uuid", nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(&mockService{}, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	svc := &mockService{
		updateFn: func(ctx context.Context, uid, did uuid.UUID, req *UpdateRequest) (*Response, error) {
			return &Response{ID: docID, Name: req.Name}, nil
		},
	}

	body, _ := json.Marshal(UpdateRequest{Name: "renamed.pdf"})
	req := httptest.NewRequest(http.MethodPatch, "/documents/"+docID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode: %v", err)
	}
	if resp.Name != "renamed.pdf" {
		t.Errorf("expected renamed.pdf, got %s", resp.Name)
	}
}

func TestHandler_Update_Forbidden(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		updateFn: func(ctx context.Context, uid, did uuid.UUID, req *UpdateRequest) (*Response, error) {
			return nil, ErrForbidden
		},
	}

	body, _ := json.Marshal(UpdateRequest{Name: "renamed.pdf"})
	req := httptest.NewRequest(http.MethodPatch, "/documents/"+uuid.NewString(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestHandler_Update_InvalidBody(t *testing.T) {
	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPatch, "/documents/"+uuid.NewString(), bytes.NewBufferString("bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(&mockService{}, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		deleteFn: func(ctx context.Context, uid, did uuid.UUID) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/documents/"+uuid.NewString(), nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		deleteFn: func(ctx context.Context, uid, did uuid.UUID) error {
			return ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/documents/"+uuid.NewString(), nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Download_Success(t *testing.T) {
	userID := uuid.New()
	docID := uuid.New()
	content := "fake pdf content"

	svc := &mockService{
		downloadFn: func(ctx context.Context, uid, did uuid.UUID) (*Document, io.ReadCloser, error) {
			return &Document{
				ID:          docID,
				Name:        "test.pdf",
				ContentType: "application/pdf",
				Size:        int64(len(content)),
			}, io.NopCloser(strings.NewReader(content)), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/documents/"+docID.String()+"/download", nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != content {
		t.Errorf("expected body %q, got %q", content, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("expected content-type application/pdf, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "test.pdf") {
		t.Errorf("expected Content-Disposition to contain filename, got %s", cd)
	}
}

func TestHandler_Download_NotFound(t *testing.T) {
	userID := uuid.New()
	svc := &mockService{
		downloadFn: func(ctx context.Context, uid, did uuid.UUID) (*Document, io.ReadCloser, error) {
			return nil, nil, ErrNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/documents/"+uuid.NewString()+"/download", nil)
	w := httptest.NewRecorder()
	newTestRouter(newTestHandler(svc, userID)).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
