package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
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

func postJSON(t *testing.T, router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("could not marshal request body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestHandler_Register_Success(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return &UserResponse{ID: uuid.New(), Email: req.Email}, nil
		},
	}

	w := postJSON(t, newTestRouter(NewHandler(svc, nil)), "/auth/register", RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

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

func TestHandler_Register_InvalidBody(t *testing.T) {
	svc := &mockService{}

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(NewHandler(svc, nil)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Register_EmailTaken(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return nil, ErrEmailTaken
		},
	}

	w := postJSON(t, newTestRouter(NewHandler(svc, nil)), "/auth/register", RegisterRequest{
		Email:    "taken@example.com",
		Password: "password123",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestHandler_Register_ServiceError(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
			return nil, errors.New("unexpected db failure")
		},
	}

	w := postJSON(t, newTestRouter(NewHandler(svc, nil)), "/auth/register", RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req *LoginRequest) (string, string, error) {
			return "access-token", "refresh-token", nil
		},
	}

	w := postJSON(t, newTestRouter(NewHandler(svc, nil)), "/auth/login", LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var hasAccess, hasRefresh bool
	for _, c := range cookies {
		if c.Name == AccessTokenCookie {
			hasAccess = true
			if !c.HttpOnly {
				t.Error("access token cookie should be HttpOnly")
			}
		}
		if c.Name == refreshTokenCookie {
			hasRefresh = true
			if !c.HttpOnly {
				t.Error("refresh token cookie should be HttpOnly")
			}
			if c.Path != "/auth/refresh" {
				t.Errorf("refresh cookie path should be /auth/refresh, got %s", c.Path)
			}
		}
	}
	if !hasAccess {
		t.Error("expected access_token cookie to be set")
	}
	if !hasRefresh {
		t.Error("expected refresh_token cookie to be set")
	}
}

func TestHandler_Login_InvalidBody(t *testing.T) {
	svc := &mockService{}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(NewHandler(svc, nil)).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req *LoginRequest) (string, string, error) {
			return "", "", ErrInvalidCredentials
		},
	}

	w := postJSON(t, newTestRouter(NewHandler(svc, nil)), "/auth/login", LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// --- Refresh ---

func TestHandler_Refresh_Success(t *testing.T) {
	svc := &mockService{
		refreshTokensFn: func(ctx context.Context, refreshToken string) (string, string, error) {
			return "new-access-token", "new-refresh-token", nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: "valid-refresh-token"})
	w := httptest.NewRecorder()
	newTestRouter(NewHandler(svc, nil)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	cookies := w.Result().Cookies()
	var hasAccess bool
	for _, c := range cookies {
		if c.Name == AccessTokenCookie {
			hasAccess = true
			if c.Value == "" {
				t.Error("expected new access token value")
			}
		}
	}
	if !hasAccess {
		t.Error("expected access_token cookie to be refreshed")
	}
}

func TestHandler_Refresh_MissingCookie(t *testing.T) {
	svc := &mockService{}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	w := httptest.NewRecorder()
	newTestRouter(NewHandler(svc, nil)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHandler_Refresh_InvalidToken(t *testing.T) {
	svc := &mockService{
		refreshTokensFn: func(ctx context.Context, refreshToken string) (string, string, error) {
			return "", "", ErrInvalidToken
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: "expired-token"})
	w := httptest.NewRecorder()
	newTestRouter(NewHandler(svc, nil)).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// cookies should be cleared on invalid refresh
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == AccessTokenCookie || c.Name == refreshTokenCookie {
			if c.MaxAge != -1 {
				t.Errorf("expected cookie %s to be cleared, MaxAge was %d", c.Name, c.MaxAge)
			}
		}
	}
}

func TestHandler_Me_Success(t *testing.T) {
	userID := uuid.New()

	router := gin.New()
	rg := router.Group("/")

	h := NewHandler(&mockService{
		getFn: func(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
			return &UserResponse{ID: id, Email: "test@example.com"}, nil
		},
	}, func(c *gin.Context) {
		// inject claims exactly as the real Authenticate middleware would
		c.Set(middleware.ContextUserKey, &token.Claims{
			UserID: userID,
			Email:  "test@example.com",
		})
		c.Next()
	})
	h.RegisterRoutes(rg)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp UserResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", resp.Email)
	}
	if resp.ID != userID {
		t.Errorf("expected %s, got %s", userID, resp.ID)
	}
}

func TestHandler_Me_Unauthorized(t *testing.T) {
	// No middleware injected — simulates missing/invalid token
	router := gin.New()
	rg := router.Group("/")
	h := NewHandler(&mockService{}, func(c *gin.Context) {
		httputil.WriteError(c, http.StatusUnauthorized, "missing access token")
		c.Abort()
	})
	h.RegisterRoutes(rg)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
