package exporter

import (
	"github.com/google/go-github/v28/github"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
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
		"created_at",
		"updated_at",
		"pushed_at",
	}
	issueLabels = []string{
		"org_name",
		"repo_name",
		"state",
		"title",
		"created_at",
		"updated_at",
		"closed_at",
		"assignee",
		"label",
	}
	pullRequestLabels = []string{
		"org_name",
		"repo_name",
		"state",
		"title",
		"created_at",
		"updated_at",
		"closed_at",
		"merged_at",
		"assignee",
		"reviewer",
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
	issueInfo = prometheus.NewDesc(
		"issue_info",
		"issue info",
		issueLabels,
		nil,
	)
	pullRequestInfo = prometheus.NewDesc(
		"pull_request_info",
		"pull request info",
		pullRequestLabels,
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
	ch <- pullRequestInfo
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
			log.Errorf("%s data not found: %v", g.org, err)
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
			log.Errorf("%s repos not found: %v", g.org, err)
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

		// set repository metrics in this loop
		for _, repo := range repos {
			if repo.GetPrivate() {
				privateCnt++
			} else {
				publicCnt++
			}
			c.setRepoMetrics(ch, repo)

			// set issue metrics in this loop
			issues, err := g.GetIssuesByRepo(repo.GetName())
			if err != nil {
				log.Errorf("%s/%s issues not found: %v", g.org, repo.GetName(), err)
			}
			for _, issue := range issues {
				c.setIssueMetrics(ch, g, repo.GetName(), issue)
			}

			// set pull request metrics in this loop
			pulls, err := g.GetPullRequestsByRepo(repo.GetName())
			if err != nil {
				log.Errorf("%s/%s pull requests not found: %v", g.org, repo.GetName(), err)
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
}

func (c *devCollector) setIssueMetrics(ch chan<- prometheus.Metric, g *GitHubCollector, repoName string, issue *github.Issue) {
	// set label string to prometheus label value
	// labelName contains multiple labels connected by comma.
	labelArr := make([]string, len(issue.Labels))
	for i, label := range issue.Labels {
		labelArr[i] = label.GetName()
	}
	labelName := strings.Join(labelArr, ",")
	labels := []string{
		g.org,
		repoName,
		issue.GetState(),
		issue.GetTitle(),
		issue.GetCreatedAt().String(),
		issue.GetUpdatedAt().String(),
		formatTime(issue.GetClosedAt()),
		issue.GetAssignee().GetLogin(),
		labelName,
	}
	ch <- prometheus.MustNewConstMetric(
		issueInfo,
		prometheus.GaugeValue,
		1.0,
		labels...,
	)
}

func (c *devCollector) setPullRequestMetrics(ch chan<- prometheus.Metric, g *GitHubCollector, repoName string, pull *github.PullRequest) {
	// set label string to prometheus label value
	// labelname contains multiple labels connected by comma.
	labelArr := make([]string, len(pull.Labels))
	for i, label := range pull.Labels {
		labelArr[i] = label.GetName()
	}
	labelName := strings.Join(labelArr, ",")

	// set reviewer string to prometheus label value
	// reviewers contains multiple reviewers connected by comma.
	reviewerArr := make([]string, len(pull.RequestedReviewers))
	for i, reviewer := range pull.RequestedReviewers {
		reviewerArr[i] = reviewer.GetLogin()
	}
	reviewers := strings.Join(reviewerArr, ",")

	labels := []string{
		g.org,
		repoName,
		pull.GetState(),
		pull.GetTitle(),
		pull.GetCreatedAt().String(),
		pull.GetUpdatedAt().String(),
		formatTime(pull.GetClosedAt()),
		formatTime(pull.GetMergedAt()),
		pull.GetAssignee().GetLogin(),
		reviewers,
		labelName,
	}
	ch <- prometheus.MustNewConstMetric(
		pullRequestInfo,
		prometheus.GaugeValue,
		1.0,
		labels...,
	)
}

// formatTime returns empty if t is zero
func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.String()
}
