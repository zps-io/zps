/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package config

import "net/url"

type RepoConfig struct {
	Priority int  `hcl:"priority"`
	Enabled  bool `hcl:"enabled"`

	Channels []string `hcl:"channels,optional"`

	Fetch   *FetchConfig   `hcl:"fetch,block"`
	Publish *PublishConfig `hcl:"publish,block"`
}

type FetchConfig struct {
	Uri       *url.URL
	UriString string `hcl:"uri"`
}

type PublishConfig struct {
	Uri       *url.URL
	UriString string `hcl:"uri"`
	Name      string `hcl:"name"`
	Prune     int    `hcl:"prune"`
}
