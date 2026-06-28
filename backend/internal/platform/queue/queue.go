package queue

import (
	"context"

	"github.com/google/uuid"
)

type Job struct {
	DocumentID uuid.UUID
}

type Queue interface {
	Push(ctx context.Context, job Job) error
	Jobs() <-chan Job
}
