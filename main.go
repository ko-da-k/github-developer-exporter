package main

import (
	"context"
	"fmt"
	"github.com/ko-da-k/github-developer-exporter/exporter"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ko-da-k/github-developer-exporter/config"
	"github.com/ko-da-k/github-developer-exporter/handlers"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// setting github client
	client, err := exporter.NewGitHubClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize github client: %v", err)
	}

	// background worker
	w := exporter.NewWorker()
	d := exporter.NewDispatcher(w)
	d.Start(ctx) // start background job queue and worker

	// setting exporter and job initialization
	orgs := strings.Split(os.Getenv("GITHUB_ORGS"), ",")
	jobs := make([]*exporter.Job, len(orgs))
	collectors := make([]*exporter.GitHubCollector, len(orgs))
	for i, org := range orgs {
		jobs[i] = exporter.NewJob(client, org)
		collectors[i] = exporter.NewGitHubCollector(org)
	}
	go func(ctx context.Context) {
		// initialized
		for _, job := range jobs {
			d.Add(job)
		}
		ticker := time.NewTicker(25 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Warnf("stop loop api call")
				return
			case <-ticker.C:
				for _, job := range jobs {
					d.Add(job)
				}
			}
		}
	}(ctx)

	// setting http server
	routes := handlers.NewRoutes()
	routes.LivenessHandler = handlers.NewLivenessHandler()
	routes.ReadinessHandler = handlers.NewReadinessHandler()
	routes.NotFoundHandler = handlers.NewNotFoundHandler()
	// custom metrics handler
	routes.MetricsHandler = handlers.NewMetricsHandler(collectors)

	handler := routes.Handler()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerConfig.Port),
		Handler: handler,
	}
	// run server
	go func() {
		log.Infof("Listen at %s port\n", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("%+v\n", err)
		}
	}()

	// graceful shutdown
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, os.Interrupt)
	log.Infof("SIGNAL %d received, then shutting down...\n", <-sigint)
	if err := server.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Warnf("Failed to gracefully shutdown: %v", err)
	}
	log.Info("Server shutdown")
}
