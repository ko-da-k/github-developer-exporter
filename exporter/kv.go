package exporter

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	// save github data for api limit.
	Kv *cache.Cache
)

func init() {
	Kv = cache.New(30*time.Minute, 35*time.Minute)
}
