package exporter

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/ko-da-k/github-developer-exporter/config"
)

type Dispatcher struct {
	workerPool chan struct{}
	jobQueue   chan *Job
	worker     Worker
	wg         sync.WaitGroup
}

func NewDispatcher(worker Worker) *Dispatcher {
	pool := make(chan struct{}, config.ServerConfig.MaxWorker)
	queue := make(chan *Job, config.ServerConfig.MaxQueue)
	return &Dispatcher{
		pool,
		queue,
		worker,
		sync.WaitGroup{},
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.wg.Add(1)
	go d.run(ctx)
}

func (d *Dispatcher) Wait() {
	d.wg.Wait()
}

func (d *Dispatcher) Add(job *Job) {
	d.jobQueue <- job
}

func (d *Dispatcher) Stop() {
	d.wg.Done()
}

func (d *Dispatcher) run(ctx context.Context) {
	wg := sync.WaitGroup{}
	// starting n number of workers
	for {
		select {
		case job := <-d.jobQueue:
			// increment the waitgroup
			wg.Add(1)
			d.workerPool <- struct{}{}

			go func(job *Job) {
				defer wg.Done()
				defer func() { <-d.workerPool }()

				log.Infof("%s job started", job.orgName)
				d.worker.Work(ctx, job)
			}(job)
		case <-ctx.Done():
			wg.Wait()
			d.Stop()
			return
		}
	}
}
