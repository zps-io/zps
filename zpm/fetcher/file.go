package fetcher

import (
	"net/url"
)

type FileFetcher struct {
	uri *url.URL
}

func NewFileFetcher(uri *url.URL) *FileFetcher {
	return &FileFetcher{uri}
}

func (f *FileFetcher) Refresh() error {
	return nil
}
