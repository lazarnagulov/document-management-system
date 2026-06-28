package document

import (
	"errors"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/middleware"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"github.com/gin-gonic/gin"
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

func (h *Handler) Get(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, ok := httputil.ParamUUID(c, "id")
	if !ok {
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

	docID, ok := httputil.ParamUUID(c, "id")
	if !ok {
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

	docID, ok := httputil.ParamUUID(c, "id")
	if !ok {
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

	docID, ok := httputil.ParamUUID(c, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), claims.UserID, docID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) List(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := pagination.New(
		httputil.QueryInt(c, "page", 1),
		httputil.QueryInt(c, "size", pagination.DefaultSize),
	)

	docs, err := h.svc.List(c.Request.Context(), claims.UserID, page)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, docs)
}

func (h *Handler) Query(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		httputil.WriteError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID, ok := httputil.ParamUUID(c, "id")
	if !ok {
		return
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.WriteError(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.svc.Query(c.Request.Context(), claims.UserID, docID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httputil.WriteError(c, http.StatusNotFound, "document not found")
	case errors.Is(err, ErrForbidden):
		httputil.WriteError(c, http.StatusForbidden, "access denied")
	case errors.Is(err, ErrUnsupportedType):
		httputil.WriteError(c, http.StatusUnprocessableEntity, "unsupported file type - use pdf, txt, or docx")
	case errors.Is(err, ErrFileTooLarge):
		httputil.WriteError(c, http.StatusRequestEntityTooLarge, "file exceeds 50MB limit")
	case errors.Is(err, ErrInvalidFile):
		httputil.WriteError(c, http.StatusBadRequest, "invalid file")
	case errors.Is(err, ErrNotReady):
		httputil.WriteError(c, http.StatusConflict, "document is still being processed")
	default:
		httputil.WriteError(c, http.StatusInternalServerError, "internal server error")
	}
}
