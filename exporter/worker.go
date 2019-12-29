package exporter

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var (
	MaxWorker = os.Getenv("MAX_WORKERS")
	MaxQueue  = os.Getenv("MAX_QUEUE")
)

type Worker struct {
	JobChannel chan Job
}

func (w Worker) Start(ctx context.Context) {
	go func() {
		for range time.Tick(25 * time.Minute) {
			select {
			case job := <-w.JobChannel:
				if err := job.Execute(ctx); err != nil {
					log.Errorf("Failed to excuse job: %v", err)
				}
			case <-ctx.Done():
				// we have received a signal to stop
				log.Warnf("worker received a cancel request")
				return
			}
		}
	}()
}
