package exporter

import (
	"github.com/google/go-github/v28/github"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
)

var (
	// resource labels
	orgLabels = []string{
		"login",
		"name",
		"url",
		"email",
		"blog",
		"created_at",
		"updated_at",
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
		"assignee",
		"label",
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
	gs []*GitHubCollector
}

func NewDevCollector(gs []*GitHubCollector) prometheus.Collector {
	return &devCollector{gs}
}

func (c *devCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- orgInfo
	ch <- orgTotalReposCount
	ch <- orgPublicReposCount
	ch <- orgPrivateReposCount
	ch <- repoInfo
	ch <- repoOpenIssueCount
	ch <- repoCollaboratorCount
	ch <- repoReleaseCount
	ch <- pullsInfo
}

func (c *devCollector) Collect(ch chan<- prometheus.Metric) {
	ok := c.collectOrgsMetrics(ch)

	// check latest query successfully
	if ok {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 1.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0.0,
		)
	}
}

// collectOrgsMetrics fetch data from cache and calculate prometheus metrics
func (c *devCollector) collectOrgsMetrics(ch chan<- prometheus.Metric) bool {
	for _, g := range c.gs {
		org, err := g.GetOrg()
		if err != nil {
			log.Errorf("%s data not found", g.org)
			return false
		}
		labels := []string{
			org.GetLogin(),
			org.GetName(),
			org.GetURL(),
			org.GetEmail(),
			org.GetBlog(),
			org.GetCreatedAt().String(),
			org.GetUpdatedAt().String(),
		}
		ch <- prometheus.MustNewConstMetric(
			orgInfo,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)

		repos, err := g.GetReposByOrg()
		if err != nil {
			log.Errorf("%s repos not found", g.org)
			return false
		}
		ch <- prometheus.MustNewConstMetric(
			orgTotalReposCount,
			prometheus.GaugeValue,
			float64(len(repos)),
			labels...,
		)
		publicCnt := 0.0
		privateCnt := 0.0
		for _, repo := range repos {
			if repo.GetPrivate() {
				privateCnt++
			} else {
				publicCnt++
			}
			c.setRepoMetrics(ch, repo)

			pulls, err := g.GetPullsByRepo(repo.GetName())
			if err != nil {
				log.Errorf("%s/%s pull requests not found", g.org, repo.GetName())
			}
			for _, pull := range pulls {
				c.setPullRequestMetrics(ch, g, repo.GetName(), pull)
			}
		}
		ch <- prometheus.MustNewConstMetric(
			orgPublicReposCount,
			prometheus.GaugeValue,
			publicCnt,
			labels...,
		)
		ch <- prometheus.MustNewConstMetric(
			orgPrivateReposCount,
			prometheus.GaugeValue,
			privateCnt,
			labels...,
		)
	}
	return true
}

func (c *devCollector) setRepoMetrics(ch chan<- prometheus.Metric, repo *github.Repository) {
	// set metrics
	labels := []string{
		repo.GetOrganization().GetLogin(),
		repo.GetName(),
		repo.GetFullName(),
		repo.GetOwner().GetLogin(),
		repo.GetURL(),
		repo.GetDefaultBranch(),
		strconv.FormatBool(repo.GetArchived()),
		repo.GetLanguage(),
		"latest_release_tag_name",
		"latest_release_at",
		repo.GetCreatedAt().String(),
		repo.GetUpdatedAt().String(),
		repo.GetPushedAt().String(),
	}
	ch <- prometheus.MustNewConstMetric(
		repoInfo,
		prometheus.GaugeValue,
		1.0,
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		repoOpenIssueCount,
		prometheus.GaugeValue,
		float64(repo.GetOpenIssuesCount()),
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		repoCollaboratorCount,
		prometheus.GaugeValue,
		1.0, // TODO
		labels...,
	)
	ch <- prometheus.MustNewConstMetric(
		repoReleaseCount,
		prometheus.GaugeValue,
		1.0, // TODO
		labels...,
	)
}

func (c *devCollector) setPullRequestMetrics(ch chan<- prometheus.Metric, g *GitHubCollector, repoName string, pull *github.PullRequest) {
	for _, label := range pull.Labels {
		labels := []string{
			g.org,
			repoName,
			pull.GetState(),
			pull.GetTitle(),
			pull.GetCreatedAt().String(),
			pull.GetUpdatedAt().String(),
			pull.GetClosedAt().String(),
			pull.GetMergedAt().String(),
			pull.GetAssignee().GetLogin(),
			label.GetName(),
		}
		ch <- prometheus.MustNewConstMetric(
			pullsInfo,
			prometheus.GaugeValue,
			1.0,
			labels...,
		)
	}
	// use "" if you would query for all pulls/
	labels := []string{
		g.org,
		repoName,
		pull.GetState(),
		pull.GetTitle(),
		pull.GetCreatedAt().String(),
		pull.GetUpdatedAt().String(),
		pull.GetClosedAt().String(),
		pull.GetMergedAt().String(),
		pull.GetAssignee().GetLogin(),
		"",
	}
	ch <- prometheus.MustNewConstMetric(
		pullsInfo,
		prometheus.GaugeValue,
		1.0,
		labels...,
	)
}
