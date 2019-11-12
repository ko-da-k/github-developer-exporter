package config

import (
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type serverConfig struct {
	Port int `default:"8888"`
}

type githubConfig struct {
	Token string `required:"true"`
	Orgs  string `required:"true"`
	// URL should be set for GitHub Enterprise
	// e.g. https://<your-domain>/api/v3/
	URL string `default:"https://api.github.com/`
}

var (
	// ServerConfig
	ServerConfig serverConfig
	// GitHubConfig
	GitHubConfig githubConfig
)

func init() {
	if err := envconfig.Process("", &ServerConfig); err != nil {
		log.Fatalf("server config error: %+v", err)
	}

	if err := envconfig.Process("GITHUB", &GitHubConfig); err != nil {
		log.Fatalf("GitHub config error: %+v", err)
	}
}
