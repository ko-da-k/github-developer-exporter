package exporter

import (
	"context"

	log "github.com/sirupsen/logrus"
)

type Worker interface {
	Work(ctx context.Context, job *Job)
}

type worker struct{}

var _ Worker = (*worker)(nil)

func NewWorker() Worker {
	return &worker{}
}

func (w *worker) Work(ctx context.Context, job *Job) {
	if err := job.Execute(ctx); err != nil {
		log.Errorf("Failed to excuse job: %v", err)
	}
}
