package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ko-da-k/github-developer-exporter/config"
	"github.com/ko-da-k/github-developer-exporter/handlers"
)

func main() {
	routes := handlers.NewRoutes()
	routes.LivenessHandler = handlers.NewLivenessHandler()
	routes.ReadinessHandler = handlers.NewReadinessHandler()
	routes.NotFoundHandler = handlers.NewNotFoundHandler()
	// custom metrics handler
	routes.MetricsHandler = handlers.NewMetricsHandler()

	handler := routes.Handler()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ServerConfig.Port),
		Handler: handler,
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		log.Warnf("Failed to gracefully shutdown: %v", err)
	}
	log.Info("Server shutdown")
}
