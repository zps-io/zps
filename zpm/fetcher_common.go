package zpm

import (
	"net/url"

	"github.com/solvent-io/zps/zps"
)

type Fetcher interface {
	Refresh() error
	Fetch(pkg *zps.Pkg) error
}

func NewFetcher(uri *url.URL, cache *Cache) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri, cache)
	default:
		return nil
	}
}