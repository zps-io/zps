package fetcher

import (
	"net/url"
)

type Fetcher interface {
	Refresh() error
}

func Get(uri *url.URL) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri)
	default:
		return nil
	}
}
