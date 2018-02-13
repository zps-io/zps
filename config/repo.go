package config

import "net/url"

type RepoConfig struct {
	Priority int
	Enabled  bool

	Fetch   *FetchConfig
	Publish *PublishConfig
}

type FetchConfig struct {
	Uri *url.URL
}

type PublishConfig struct {
	Uri   *url.URL
	Name  string
	Prune int
}
