package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func RecordMetrics() {
	c, err := NewDevCollector()
	if err != nil {
		log.Errorf("Collector Initialization Error: %v", err)
		return
	}
	prometheus.MustRegister(c)
	return
}
