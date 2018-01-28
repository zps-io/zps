package fetcher

import (
	"net/url"
	"github.com/solvent-io/zps/zps"
)

type Fetcher interface {
	Refresh() error
	Fetch(pkg *zps.Pkg) error
}

func Get(uri *url.URL, cachePath string) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri, cachePath)
	default:
		return nil
	}
}
