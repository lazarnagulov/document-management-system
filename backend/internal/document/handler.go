package document

import (
	"errors"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc  Service
	auth gin.HandlerFunc
}

func NewHandler(svc Service, auth gin.HandlerFunc) *Handler {
	return &Handler{svc: svc, auth: auth}
}

func (h *Handler) Create(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)
	header, err := c.FormFile("document")
	if err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "document file is required")
		return
	}

	file, err := header.Open()
	if err != nil {
		httputil.WriteError(c, http.StatusInternalServerError, "failed to read file")
		return
	}
	defer file.Close()

	resp, err := h.svc.Create(c.Request.Context(), claims.UserID, header, file)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) List(c *gin.Context) {

}

func (h *Handler) Get(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "invalid document id")
		return
	}

	doc, err := h.svc.Get(c.Request.Context(), claims.UserID, docID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *Handler) Update(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "invalid document id")
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	doc, err := h.svc.Update(c.Request.Context(), claims.UserID, docID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, doc)
}

func (h *Handler) Download(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "invalid document id")
		return
	}

	doc, reader, err := h.svc.Download(c.Request.Context(), claims.UserID, docID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	defer reader.Close()

	c.DataFromReader(http.StatusOK, doc.Size, doc.ContentType, reader, map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, doc.Name),
	})
}

func (h *Handler) Delete(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.WriteError(c, http.StatusBadRequest, "invalid document id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), claims.UserID, docID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(c, http.StatusNotFound, "document not found")
	case errors.Is(err, ErrForbidden):
		httputil.WriteError(c, http.StatusForbidden, "access denied")
	case errors.Is(err, ErrUnsupportedType):
		httputil.WriteError(c, http.StatusUnprocessableEntity, "unsupported file type — use pdf, txt, or docx")
	case errors.Is(err, ErrFileTooLarge):
		httputil.WriteError(c, http.StatusRequestEntityTooLarge, "file exceeds 50MB limit")
	case errors.Is(err, ErrInvalidFile):
		httputil.WriteError(c, http.StatusBadRequest, "invalid file")
	default:
		httputil.WriteError(c, http.StatusInternalServerError, "internal server error")
	}
}
