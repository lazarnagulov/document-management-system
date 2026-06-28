package queue

import (
	"context"
	"fmt"
)

type MemoryQueue struct {
	jobs chan Job
}

func NewMemoryQueue(buffer int) *MemoryQueue {
	return &MemoryQueue{jobs: make(chan Job, buffer)}
}

func (q *MemoryQueue) Push(ctx context.Context, job Job) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("push cancelled: %w", ctx.Err())
	default:
		return fmt.Errorf("queue full, document %s will not be processed", job.DocumentID)
	}
}

func (q *MemoryQueue) Jobs() <-chan Job {
	return q.jobs
}
