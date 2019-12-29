package exporter

import (
	"context"
	"sync"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type config struct {
	MaxWorker int `default:"2"`
	MaxQueue  int `default:"5"`
}

var (
	Config config
)

func init() {
	if err := envconfig.Process("", &Config); err != nil {
		log.Errorf("failed to load environment variable")
	}
}

type Dispatcher struct {
	workerPool chan struct{}
	jobQueue   chan *Job
	worker     Worker
	wg         sync.WaitGroup
}

func NewDispatcher(worker Worker) *Dispatcher {
	pool := make(chan struct{}, Config.MaxWorker)
	queue := make(chan *Job, Config.MaxQueue)
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
