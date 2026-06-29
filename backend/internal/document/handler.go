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

// @Summary      Upload a document
// @Description  Uploads a new document (PDF, TXT, or DOCX). Max file size is 50MB. The document is queued for processing after upload; poll the
//
//	status field until it reads READY before querying.
//
// @Tags         documents
// @Accept       multipart/form-data
// @Produce      json
// @Param        document  formData  file            true  "Document file (pdf, txt, docx)"
// @Success      201       {object}  document.Response
// @Failure      400       {object}  httputil.ErrorResponse  "Missing or invalid file"
// @Failure      401       {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      413       {object}  httputil.ErrorResponse  "File exceeds 50MB limit"
// @Failure      422       {object}  httputil.ErrorResponse  "Unsupported file type"
// @Failure      500       {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/ [post]
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

// @Summary      Get a document
// @Description  Returns metadata for a single document owned by the authenticated user.
// @Tags         documents
// @Produce      json
// @Param        id   path      string  true  "Document UUID"
// @Success      200  {object}  document.Response
// @Failure      401  {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  httputil.ErrorResponse  "Access denied"
// @Failure      404  {object}  httputil.ErrorResponse  "Document not found"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/{id} [get]
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

// @Summary      Rename a document
// @Description  Updates the name of an existing document owned by the authenticated user.
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Document UUID"
// @Param        body  body      document.UpdateRequest  true  "Updated fields"
// @Success      200   {object}  document.Response
// @Failure      400   {object}  httputil.ErrorResponse  "Invalid request body"
// @Failure      401   {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  httputil.ErrorResponse  "Access denied"
// @Failure      404   {object}  httputil.ErrorResponse  "Document not found"
// @Failure      500   {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/{id} [patch]
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

// @Summary      Download a document
// @Description  Streams the original file as an attachment. The Content-Disposition header will contain the original filename.
// @Tags         documents
// @Produce      application/octet-stream
// @Param        id  path  string  true  "Document UUID"
// @Success      200  {file}    binary
// @Failure      401  {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  httputil.ErrorResponse  "Access denied"
// @Failure      404  {object}  httputil.ErrorResponse  "Document not found"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/{id}/download [get]
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

// @Summary      Delete a document
// @Description  Permanently deletes a document and its stored file. This action is irreversible.
// @Tags         documents
// @Param        id  path  string  true  "Document UUID"
// @Success      204
// @Failure      401  {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  httputil.ErrorResponse  "Access denied"
// @Failure      404  {object}  httputil.ErrorResponse  "Document not found"
// @Failure      500  {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/{id} [delete]
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

// @Summary      List documents
// @Description  Returns a paginated list of documents owned by the authenticated user, ordered by upload date descending.
// @Tags         documents
// @Produce      json
// @Param        page  query     int  false  "Page number (default: 1)"
// @Param        size  query     int  false  "Page size (default: 15)"
// @Success      200   {object}  pagination.Response[document.Response]
// @Failure      401   {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      500   {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/ [get]
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

// @Summary      Query a document with AI
// @Description  Asks a natural language question about the document content using semantic search and an LLM. The document status must
//
//	be READY; a 409 is returned while it is still processing.
//
// @Tags         documents
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Document UUID"
// @Param        body  body      document.QueryRequest   true  "Question payload"
// @Success      200   {object}  document.QueryResponse
// @Failure      400   {object}  httputil.ErrorResponse  "Invalid request body"
// @Failure      401   {object}  httputil.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  httputil.ErrorResponse  "Access denied"
// @Failure      404   {object}  httputil.ErrorResponse  "Document not found"
// @Failure      409   {object}  httputil.ErrorResponse  "Document is still processing"
// @Failure      500   {object}  httputil.ErrorResponse  "Internal server error"
// @Security     AccessTokenCookie
// @Router       /documents/{id}/query [post]
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
