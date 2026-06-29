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

// @Summary      Log in a user
// @Description  Authenticates a user with email and password. On success, sets HttpOnly cookies containing the access and refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login credentials"
// @Success      200
// @Failure      400  {object}  httputil.ErrorResponse  "Invalid request body"
// @Failure      401  {object}  httputil.ErrorResponse  "Invalid credentials"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Security     RefreshTokenCookie
// @Router       /auth/login [post]
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

// Refresh godoc
// @Summary      Refresh authentication tokens
// @Description  Reads the refresh token from the HttpOnly cookie, validates it, and issues a new access/refresh token pair via cookies. Clears cookies on failure.
// @Tags         auth
// @Produce      json
// @Success      200
// @Failure      401  {object}  httputil.ErrorResponse  "Missing or invalid refresh token"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Router       /auth/refresh [post]
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

// @Summary      Log out the current user
// @Description  Clears the access and refresh token cookies, effectively logging the user out.
// @Tags         auth
// @Success      204
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	clearAuthCookies(c)
	c.Status(http.StatusNoContent)
}

// @Summary      Register a new user
// @Description  Creates a new user account. Returns the created user profile on success.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Registration details"
// @Success      201  {object}  UserResponse
// @Failure      400  {object}  httputil.ErrorResponse  "Invalid request body"
// @Failure      409  {object}  httputil.ErrorResponse  "Email already in use"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Router       /auth/register [post]
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

// @Summary      Get current user
// @Description  Returns the profile of the currently authenticated user, identified via the access token cookie.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  UserResponse
// @Failure      401  {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      404  {object}  httputil.ErrorResponse  "User not found"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /auth/me [get]
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
