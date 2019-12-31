package exporter

import (
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/ko-da-k/github-developer-exporter/config"
)

var (
	// save github data for api limit.
	Kv *cache.Cache
)

func init() {
	Kv = cache.New(
		// add more 3 minutes to prevent nil response
		time.Duration(config.GitHubConfig.Interval+3)*time.Minute,
		time.Duration(config.GitHubConfig.Interval+3)*time.Minute,
	)
}
