package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ko-da-k/github-developer-exporter/exporter"
)

func NewMetricsHandler(gs []*exporter.GitHubCollector) http.Handler {
	exporter.RecordMetrics(gs)

	return promhttp.Handler()
}
