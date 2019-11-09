package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	repoStatus = prometheus.NewDesc(
		"repository_info",
		"repository info",
		[]string{
			"repo_name",
			"org_name",
			"created_at",
			"updated_at",
			"last_release",
			"total_issue_count",
			"open_issue_count",
			"close_issue_count",
			"total_pr_count",
			"open_pr_count",
			"merged_pr_count",
			"close_pr_count",
			"contributor_count",
		},
		nil,
	)
)

type devCollector struct {
	repo *prometheus.Desc
}

func NewDevCollector() prometheus.Collector {
	return &devCollector{
		repoStatus,
	}
}

func (c *devCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.repo
}

func (c *devCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		c.repo,
		prometheus.GaugeValue,
		1,
		"repo_name",
		"org_name",
		"created_at",
		"updated_at",
		"last_release",
		"total_issue_count",
		"open_issue_count",
		"close_issue_count",
		"total_pr_count",
		"open_pr_count",
		"merged_pr_count",
		"close_pr_count",
		"contributor_count",
	)
}
