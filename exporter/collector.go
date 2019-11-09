package exporter

import (
	"context"

	"github.com/google/go-github/v28/github"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
)

var (
	// resource labels
	orgLabels = []string{
		"login",
		"name",
		"url",
		"email",
		"blog",
		"has_organization_projects",
		"has_repository_projects",
		"created_at",
		"upated_at",
	}
	repoLabels = []string{
		"org_name",
		"name",
		"full_name",
		"owner",
		"url",
		"default_branch",
		"archived",
		"language",
		"latest_release_tag_name",
		"latest_released_at",
		"created_at",
		"updated_at",
		"pushed_at",
	}
	pullsLabels = []string{
		"org_name",
		"repo_name",
		"state",
		"title",
		"created_at",
		"updated_at",
		"closed_at",
		"merged_at",
		"asignee",
		"reviewer",
	}

	// prometheus description
	up = prometheus.NewDesc(
		"up",
		"Was the last query successful.",
		nil,
		nil,
	)
	orgInfo = prometheus.NewDesc(
		"org_info",
		"organization info",
		orgLabels,
		nil,
	)
	orgTotalReposCount = prometheus.NewDesc(
		"org_total_repos_count",
		"How many repositories are in the organization.",
		orgLabels,
		nil,
	)
	orgPublicReposCount = prometheus.NewDesc(
		"org_public_repos_count",
		"How many public repositories are in the organization.",
		orgLabels,
		nil,
	)
	orgPrivateReposCount = prometheus.NewDesc(
		"org_private_repos_count",
		"How many private repositories are in the organization.",
		orgLabels,
		nil,
	)
	orgTotalGistCount = prometheus.NewDesc(
		"org_total_gist_count",
		"How many gists are in the organization.",
		orgLabels,
		nil,
	)
	orgPublicGistCount = prometheus.NewDesc(
		"org_public_gist_count",
		"How many public gists are in the organization.",
		orgLabels,
		nil,
	)
	orgPrivateGistCount = prometheus.NewDesc(
		"org_private_gist_count",
		"How many private gists are in the organization.",
		orgLabels,
		nil,
	)
	orgMemberCount = prometheus.NewDesc(
		"org_member_count",
		"How many private gists are in the organization.",
		orgLabels,
		nil,
	)
	repoInfo = prometheus.NewDesc(
		"repo_info",
		"repository info",
		repoLabels,
		nil,
	)
	repoOpenIssueCount = prometheus.NewDesc(
		"repo_open_issue_count",
		"How many open issues are in the repository.",
		repoLabels,
		nil,
	)
	repoCollaboratorCount = prometheus.NewDesc(
		"repo_collaborator_count",
		"How many collaborators are in the repository.",
		repoLabels,
		nil,
	)
	repoReleaseCount = prometheus.NewDesc(
		"repo_release_count",
		"How many releases are in the repository.",
		repoLabels,
		nil,
	)
	pullsInfo = prometheus.NewDesc(
		"pulls_info",
		"pulls info",
		pullsLabels,
		nil,
	)
)

type devCollector struct {
	client *github.Client
	orgs   []string
}

// NewGitHubClient returns
func NewGitHubClient(ctx context.Context, token string, org string, baseURL string, uploadURL string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	// TODO: because I mainly use gh:e
	client, err := github.NewEnterpriseClient(baseURL, uploadURL, tc)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewDevCollector() prometheus.Collector {
	return &devCollector{
		&github.Client{},
		[]string{"octocat"},
	}
}

func (c *devCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- orgInfo
	ch <- orgTotalReposCount
	ch <- orgPublicReposCount
	ch <- orgPrivateReposCount
	ch <- orgTotalGistCount
	ch <- orgPublicGistCount
	ch <- orgPrivateGistCount
	ch <- orgMemberCount
	ch <- repoInfo
	ch <- repoOpenIssueCount
	ch <- repoCollaboratorCount
	ch <- repoReleaseCount
	ch <- pullsInfo
}

func (c *devCollector) Collect(ch chan<- prometheus.Metric) {
	ok := c.collectOrgsMetrics(ch)
	ok = c.collectReposMetrics(ch) && ok
	ok = c.collectPullsMetrics(ch) && ok

	// check latest query successfully
	if ok {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 1,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
	}
}

func (c *devCollector) collectOrgsMetrics(ch chan<- prometheus.Metric) bool {
	ch <- prometheus.MustNewConstMetric(
		orgInfo,
		prometheus.GaugeValue,
		1,
		"login",
		"name",
		"url",
		"email",
		"blog",
		"has_organization_projects",
		"has_repository_projects",
		"created_at",
		"upated_at",
	)
	return true
}

func (c *devCollector) collectReposMetrics(ch chan<- prometheus.Metric) bool {
	ch <- prometheus.MustNewConstMetric(
		repoInfo,
		prometheus.GaugeValue,
		1,
		"org_name",
		"name",
		"full_name",
		"owner",
		"url",
		"default_branch",
		"archived",
		"language",
		"latest_release_tag_name",
		"latest_released_at",
		"created_at",
		"updated_at",
		"pushed_at",
	)
	return true
}

func (c *devCollector) collectPullsMetrics(ch chan<- prometheus.Metric) bool {
	ch <- prometheus.MustNewConstMetric(
		pullsInfo,
		prometheus.GaugeValue,
		1,
		"org_name",
		"repo_name",
		"state",
		"title",
		"created_at",
		"updated_at",
		"closed_at",
		"merged_at",
		"asignee",
		"reviewer",
	)
	return true
}
