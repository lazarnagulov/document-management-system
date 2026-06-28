package processing

import (
	"context"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/queue"
)

type Worker struct {
	queue     queue.Queue
	processor *Processor
}

func NewWorker(queue queue.Queue, processor *Processor) *Worker {
	return &Worker{queue: queue, processor: processor}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case job, ok := <-w.queue.Jobs():
				if !ok {
					return
				}
				w.processor.Process(ctx, job.DocumentID)
			case <-ctx.Done():
				return
			}
		}
	}()
}
