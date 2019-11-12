package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ko-da-k/github-developer-exporter/exporter"
)

func NewMetricsHandler() http.Handler {
	exporter.RecordMetrics()

	return promhttp.Handler()
}
