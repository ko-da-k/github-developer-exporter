package exporter

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/patrickmn/go-cache"
)

type Job struct {
	client  *github.Client
	orgName string
}

func (j *Job) Execute(ctx context.Context) error {
	if err := j.setCacheByOrg(ctx); err != nil {
		return fmt.Errorf("failed to set %s org: %w", j.orgName, err)
	}
	return nil
}

func (j *Job) setCacheByOrg(ctx context.Context) error {
	org, _, err := j.client.Organizations.Get(ctx, j.orgName)
	if _, ok := err.(*github.RateLimitError); ok {
		return fmt.Errorf("access Rate Limit: %w", err)
	} else if err != nil {
		return fmt.Errorf("failed to get %s org: %w", j.orgName, err)
	}
	// send object to global cache Kv
	Kv.Set(j.orgName, org, cache.DefaultExpiration)
	// fetch repositories in the organization
	repoOption := &github.RepositoryListByOrgOptions{
		Type:        "all",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := j.client.Repositories.ListByOrg(ctx, j.orgName, repoOption)
		if err != nil {
			return fmt.Errorf("failed to fetch repos in %s org: %w", j.orgName, err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		repoOption.Page = resp.NextPage
	}
	// send object to global cache Kv
	Kv.Set(fmt.Sprintf("%s-repos", j.orgName), allRepos, cache.DefaultExpiration)

	// fet members in the organization
	memberOption := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allMembers []*github.User
	for {
		members, resp, err := j.client.Organizations.ListMembers(ctx, j.orgName, memberOption)
		if err != nil {
			return fmt.Errorf("failed to fetch members: %w", err)
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		memberOption.Page = resp.NextPage
	}
	Kv.Set(fmt.Sprintf("%s-members", j.orgName), allMembers, cache.DefaultExpiration)
	return nil
}

func (j *Job) setCacheByRepo(ctx context.Context) error {
	// read repositories from cache
	ri, found := Kv.Get(fmt.Sprintf("%s-members", j.orgName))
	repos, ok := ri.([]*github.Repository)
	if !found || !ok {
		return fmt.Errorf("failed to read repositories from cache")
	}
	// fetch pull requests in the repository
	prListOption := &github.PullRequestListOptions{
		State:     "all",
		Head:      "",
		Base:      "",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 20, // Limited
		},
	}
	for _, repo := range repos {
		pulls, _, err := j.client.PullRequests.List(ctx, j.orgName, repo.GetName(), prListOption)
		if _, ok := err.(*github.RateLimitError); ok {
			return fmt.Errorf("Access Rate Limit: %w", err)
		} else if err != nil {
			return fmt.Errorf("Failed to fetch %s pulls: %w", repo.GetName(), err)
		}
		Kv.Set(fmt.Sprintf("%s-%s-pulls", j.orgName, repo.GetName()), pulls, cache.DefaultExpiration)
	}
	return nil
}
