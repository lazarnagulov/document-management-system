package document

import (
	"time"

	"github.com/google/uuid"
)

type DocumentStatus string

const (
	StatusUploaded   DocumentStatus = "UPLOADED"
	StatusProcessing DocumentStatus = "PROCESSING"
	StatusReady      DocumentStatus = "READY"
	StatusFailed     DocumentStatus = "FAILED"
)

type Chunk struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Content    string
	Embedding  []float32
	ChunkIndex int
	CreatedAt  time.Time
}

type Document struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	ContentType string
	Size        int64
	Status      DocumentStatus
	StorageKey  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateRequest struct {
	Name string
}

type UpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

type Response struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	ContentType string         `json:"content_type"`
	Size        int64          `json:"size"`
	Status      DocumentStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type QueryRequest struct {
	Question string `json:"question" binding:"required"`
}

type QueryResponse struct {
	Answer   string `json:"answer"`
	Question string `json:"question"`
}

func ToResponse(d *Document) *Response {
	return &Response{
		ID:          d.ID,
		Name:        d.Name,
		ContentType: d.ContentType,
		Status:      d.Status,
		Size:        d.Size,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToResponseList(docs []*Document) []*Response {
	result := make([]*Response, len(docs))
	for i, d := range docs {
		result[i] = ToResponse(d)
	}
	return result
}
