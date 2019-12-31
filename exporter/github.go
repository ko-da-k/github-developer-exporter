package exporter

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"

	"github.com/ko-da-k/github-developer-exporter/config"
)

type GitHubCollector struct {
	org string
}

type collector interface {
	GetOrg() (*github.Organization, error)
	GetReposByOrg() ([]*github.Repository, error)
	GetIssuesByRepo(repoName string) ([]*github.Issue, error)
	GetPullRequestsByRepo(repoName string) ([]*github.PullRequest, error)
}

var _ collector = (*GitHubCollector)(nil)

func NewGitHubCollector(org string) *GitHubCollector {
	return &GitHubCollector{org}
}

func (g *GitHubCollector) GetOrg() (*github.Organization, error) {
	oi, found := Kv.Get(g.org)
	org, ok := oi.(*github.Organization)
	if !found {
		return nil, fmt.Errorf("%s not found", g.org)
	}
	if !ok {
		return nil, fmt.Errorf("type conversion failed")
	}
	return org, nil
}

func (g *GitHubCollector) GetReposByOrg() ([]*github.Repository, error) {
	rsi, found := Kv.Get(fmt.Sprintf("%s-repos", g.org))
	repos, ok := rsi.([]*github.Repository)
	if !found {
		return nil, fmt.Errorf("%s repos not found in cache", g.org)
	}
	if !ok {
		return nil, fmt.Errorf("type conversion failed")
	}
	return repos, nil
}

func (g *GitHubCollector) GetIssuesByRepo(repoName string) ([]*github.Issue, error) {
	ii, found := Kv.Get(fmt.Sprintf("%s-%s-issues", g.org, repoName))
	issues, ok := ii.([]*github.Issue)
	if !found {
		return nil, fmt.Errorf("%s/%s issues not found in cache", g.org, repoName)
	}
	if !ok {
		return nil, fmt.Errorf("type conversion failed")
	}
	return issues, nil
}

func (g *GitHubCollector) GetPullRequestsByRepo(repoName string) ([]*github.PullRequest, error) {
	psi, found := Kv.Get(fmt.Sprintf("%s-%s-pulls", g.org, repoName))
	pulls, ok := psi.([]*github.PullRequest)
	if !found {
		return nil, fmt.Errorf("%s/%s pull requests not found in cache", g.org, repoName)
	}
	if !ok {
		return nil, fmt.Errorf("type conversion failed")
	}
	return pulls, nil
}

// NewGitHubClient constructor
func NewGitHubClient(ctx context.Context) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: config.GitHubConfig.Token,
		},
	)
	tc := oauth2.NewClient(ctx, ts)

	// TODO: because I mainly use gh:e
	client, err := github.NewEnterpriseClient(config.GitHubConfig.URL, config.GitHubConfig.URL, tc)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize github client: %w", err)
	}
	return client, nil
}
