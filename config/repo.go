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
	Priority int
	Enabled  bool

	Channels []string

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
