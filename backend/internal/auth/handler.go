package auth

import (
	"errors"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc  Service
	auth gin.HandlerFunc
}

func NewHandler(s Service, auth gin.HandlerFunc) *Handler {
	return &Handler{svc: s, auth: auth}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	accessToken, refreshToken, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	setAuthCookies(c, accessToken, refreshToken)
	c.Status(http.StatusOK)
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		httputil.WriteError(c, http.StatusUnauthorized, "missing refresh token")
		return
	}

	accessToken, newRefreshToken, err := h.svc.RefreshTokens(c.Request.Context(), refreshToken)
	if err != nil {
		clearAuthCookies(c)
		h.handleError(c, err)
		return
	}

	setAuthCookies(c, accessToken, newRefreshToken)
	c.Status(http.StatusOK)
}

func (h *Handler) Logout(c *gin.Context) {
	clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Me(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return
	}

	resp, err := h.svc.Get(c.Request.Context(), claims.UserID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		httputil.WriteError(c, http.StatusConflict, "email already in use")
	case errors.Is(err, ErrInvalidCredentials):
		httputil.WriteError(c, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, ErrInvalidToken):
		httputil.WriteError(c, http.StatusUnauthorized, "invalid or expired token")
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(c, http.StatusNotFound, "not found")
	default:
		httputil.WriteError(c, http.StatusInternalServerError, "internal server error")
	}
}
