package job

import (
	"context"
	"log"
)

type PublishJob struct {
	SessionID string
}

type PublishHandler interface {
	HandlePublish(context.Context, PublishJob) error
}

type Runner struct {
	logger  *log.Logger
	handler PublishHandler
	queue   chan PublishJob
}

func NewRunner(logger *log.Logger, handler PublishHandler) *Runner {
	if logger == nil {
		logger = log.Default()
	}
	return &Runner{
		logger:  logger,
		handler: handler,
		queue:   make(chan PublishJob, 64),
	}
}

func (r *Runner) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-r.queue:
				if err := r.handler.HandlePublish(ctx, job); err != nil {
					r.logger.Printf("publish job failed for session %s: %v", job.SessionID, err)
				}
			}
		}
	}()
}

func (r *Runner) Enqueue(job PublishJob) {
	r.queue <- job
}
