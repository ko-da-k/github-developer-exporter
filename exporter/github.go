package exporter

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2"
	"os"
)

var (
	Kv *cache.Cache // save github data for api limit.
)

type GitHubCollector struct {
	org string
}

type collector interface {
	GetOrg() (*github.Organization, error)
	GetReposByOrg() ([]*github.Repository, error)
	GetPullsByRepo(repoName string) ([]*github.PullRequest, error)
}

var _ collector = (*GitHubCollector)(nil)

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
	rsi, found := Kv.Get(fmt.Sprintf("%s-repos"))
	repos, ok := rsi.([]*github.Repository)
	if !found {
		return nil, fmt.Errorf("%s repos not found in cache", g.org)
	}
	if !ok {
		return nil, fmt.Errorf("type conversion failed")
	}
	return repos, nil
}

func (g *GitHubCollector) GetPullsByRepo(repoName string) ([]*github.PullRequest, error) {
	psi, found := Kv.Get(fmt.Sprintf("%s-%s-pulls"))
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
