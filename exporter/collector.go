package exporter

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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
	gcache *cache.Cache
}

// NewGitHubClient returns
func newGitHubClient(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	url := os.Getenv("GITHUB_URL")

	if token == "" {
		return nil, fmt.Errorf("token should be set in GITHIB_TOKEN environment value")
	}
	if url == "" {
		url = "https://api.github.com/"
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	// TODO: because I mainly use gh:e
	client, err := github.NewEnterpriseClient(url, url, tc)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewDevCollector() (prometheus.Collector, error) {
	ctx := context.Background()
	c, err := newGitHubClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("client initialization error: %w", err)
	}

	orgs := strings.Split(os.Getenv("GITHUB_ORGS"), ",")

	if orgs[0] == "" {
		return nil, fmt.Errorf("organization name shoud be set in GITHUB_ORGS")
	}

	// cache
	ch := cache.New(30*time.Minute, 45*time.Minute)

	return &devCollector{
		c,
		orgs,
		ch,
	}, nil
}

func (c *devCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- orgInfo
	ch <- orgTotalReposCount
	ch <- orgPublicReposCount
	ch <- orgPrivateReposCount
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
			up, prometheus.GaugeValue, 1.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0.0,
		)
	}
}

func (c *devCollector) collectOrgsMetrics(ch chan<- prometheus.Metric) bool {
	ctx := context.Background()
	for _, orgName := range c.orgs {
		org, _, err := c.client.Organizations.Get(ctx, orgName)
		if _, ok := err.(*github.RateLimitError); ok {
			log.Errorf("Access Rate Limit: %v", err)
			return false
		} else if err != nil {
			log.Errorf("Failed to get %s org: %v", orgName, err)
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

		repoOption := &github.RepositoryListByOrgOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		var allRepos []*github.Repository
		ri, found := c.gcache.Get("repos")
		allRepos, ok := ri.([]*github.Repository)
		if !found || !ok {
			for {
				repos, resp, err := c.client.Repositories.ListByOrg(ctx, orgName, repoOption)
				if err != nil {
					log.Errorf("failed to fetch repos: %v", err)
					return false
				}
				allRepos = append(allRepos, repos...)
				if resp.NextPage == 0 {
					break
				}
				repoOption.Page = resp.NextPage

			}
			c.gcache.Set("repos", allRepos, cache.DefaultExpiration)
		}

		ch <- prometheus.MustNewConstMetric(
			orgTotalReposCount,
			prometheus.GaugeValue,
			float64(len(allRepos)),
			labels...,
		)
		publicCnt := 0.0
		privateCnt := 0.0
		for _, repo := range allRepos {
			if repo.GetPrivate() {
				privateCnt++
			} else {
				publicCnt++
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

		// fetch members list
		memberOption := &github.ListMembersOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		var allMembers []*github.User
		ui, found := c.gcache.Get("members")
		allMembers, ok = ui.([]*github.User)
		if !found || !ok {
			for {
				members, resp, err := c.client.Organizations.ListMembers(ctx, orgName, memberOption)
				if err != nil {
					log.Errorf("failed to fetch members: %v", err)
					return false
				}
				allMembers = append(allMembers, members...)
				if resp.NextPage == 0 {
					break
				}
				memberOption.Page = resp.NextPage

			}
			c.gcache.Set("users", allMembers, cache.DefaultExpiration)
		}
		ch <- prometheus.MustNewConstMetric(
			orgMemberCount,
			prometheus.GaugeValue,
			float64(len(allMembers)),
			labels...,
		)
	}
	return true
}

func (c *devCollector) collectReposMetrics(ch chan<- prometheus.Metric) bool {
	ri, found := c.gcache.Get("repos")
	allRepos, ok := ri.([]*github.Repository)
	if !found || !ok {
		log.Errorf("failed to fetch repos from cache")
		return false
	}
	for _, repo := range allRepos {
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
	return true
}

func (c *devCollector) collectPullsMetrics(ch chan<- prometheus.Metric) bool {
	isSuccess := true
	ctx := context.Background()
	ri, found := c.gcache.Get("repos")
	allRepos, ok := ri.([]*github.Repository)
	if !found || !ok {
		log.Errorf("failed to fetch repos from cache")
		return false
	}
	prListOption := &github.PullRequestListOptions{
		State:     "all",
		Head:      "",
		Base:      "",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 15, // Limited
		},
	}
	for _, repo := range allRepos {
		owner := repo.GetOwner()
		orgName := owner.GetLogin()
		repoName := repo.GetName()

		var pulls []*github.PullRequest
		pri, found := c.gcache.Get(fmt.Sprintf("%s-pulls", repoName))
		pulls, ok = pri.([]*github.PullRequest)
		if !found || !ok {
			pulls, _, err := c.client.PullRequests.List(ctx, orgName, repoName, prListOption)
			if _, ok := err.(*github.RateLimitError); ok {
				log.Errorf("Access Rate Limit: %v", err)
				return false
			} else if err != nil {
				log.Errorf("Failed to fetch %s pulls: %v", repo.GetName(), err)
				isSuccess = false
			}
			c.gcache.Set(fmt.Sprintf("%s-pulls", repoName), pulls, cache.DefaultExpiration)
		}
		for _, pull := range pulls {
			for _, label := range pull.Labels {
				labels := []string{
					orgName,
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
				owner.GetLogin(),
				repo.GetName(),
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
	}
	return isSuccess
}
