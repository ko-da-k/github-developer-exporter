package exporter

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
	repoGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "repository_count",
			Help: "Number of repositories",
		},
		[]string{"device"},
	)
)

func RecordMetrics() {
	prometheus.MustRegister(repoGauge)
	repoGauge.WithLabelValues("deviceA").Set(10)
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}
