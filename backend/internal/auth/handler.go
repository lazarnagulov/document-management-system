package auth

import (
	"errors"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(s Service) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) Login(c *gin.Context) {

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

func (h *Handler) Logout(c *gin.Context) {

}

func (h *Handler) Me(c *gin.Context) {

}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailTaken):
		httputil.WriteError(c, http.StatusConflict, "email already in use")
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(c, http.StatusNotFound, "user not found")
	case errors.Is(err, ErrInvalidRequest):
		httputil.WriteError(c, http.StatusBadRequest, "invalid request")
	default:
		httputil.WriteError(c, http.StatusInternalServerError, "internal server error")
	}
}
