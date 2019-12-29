package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

func RecordMetrics(gs []*GitHubCollector) {
	c := NewDevCollector(gs)
	prometheus.MustRegister(c)
	return
}
