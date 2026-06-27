package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestRegister_Success(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return &UserResponse{ID: uuid.New(), Email: req.Email}, nil
		},
	}

	body, _ := json.Marshal(RegisterRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(NewHandler(svc)).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", resp.Email)
	}
}

func TestRegister_InvalidBody(t *testing.T) {
	svc := &mockService{}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(NewHandler(svc)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return nil, ErrEmailTaken
		},
	}

	body, _ := json.Marshal(RegisterRequest{Email: "taken@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(NewHandler(svc)).ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRegister_ServiceError(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return nil, errors.New("unexpected db failure")
		},
	}

	body, _ := json.Marshal(RegisterRequest{Email: "test@example.com", Password: "password123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	newTestRouter(NewHandler(svc)).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
