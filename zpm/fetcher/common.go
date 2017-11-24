package fetcher

import (
	"net/url"
)

type Fetcher interface {
	Refresh() error
}

func Get(uri *url.URL, cachePath string) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri, cachePath)
	default:
		return nil
	}
}
