package job

import (
	"context"
	"log"
)

type PublishJob struct {
	SessionID string
}

type PipelineRunJob struct {
	RunID string
}

type PublishHandler interface {
	HandlePublish(context.Context, PublishJob) error
}

type PipelineRunHandler interface {
	HandlePipelineRun(context.Context, PipelineRunJob) error
}

type Runner struct {
	logger          *log.Logger
	publishHandler  PublishHandler
	pipelineHandler PipelineRunHandler
	publishQueue    chan PublishJob
	pipelineQueue   chan PipelineRunJob
}

func NewRunner(logger *log.Logger, publishHandler PublishHandler, pipelineHandler PipelineRunHandler) *Runner {
	if logger == nil {
		logger = log.Default()
	}
	return &Runner{
		logger:          logger,
		publishHandler:  publishHandler,
		pipelineHandler: pipelineHandler,
		publishQueue:    make(chan PublishJob, 64),
		pipelineQueue:   make(chan PipelineRunJob, 64),
	}
}

func (r *Runner) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-r.publishQueue:
				if r.publishHandler == nil {
					continue
				}
				if err := r.publishHandler.HandlePublish(ctx, job); err != nil {
					r.logger.Printf("publish job failed for session %s: %v", job.SessionID, err)
				}
			case job := <-r.pipelineQueue:
				if r.pipelineHandler == nil {
					continue
				}
				if err := r.pipelineHandler.HandlePipelineRun(ctx, job); err != nil {
					r.logger.Printf("pipeline run job failed for run %s: %v", job.RunID, err)
				}
			}
		}
	}()
}

func (r *Runner) Enqueue(job PublishJob) {
	r.publishQueue <- job
}

func (r *Runner) EnqueuePipelineRun(job PipelineRunJob) {
	r.pipelineQueue <- job
}
