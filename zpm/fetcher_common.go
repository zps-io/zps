/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zpm

import (
	"net/url"

	"github.com/fezz-io/zps/zps"
)

type Fetcher interface {
	Refresh() error
	Fetch(pkg *zps.Pkg) error
}

func NewFetcher(uri *url.URL, cache *Cache) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri, cache)
	case "https":
		return NewHttpsFetcher(uri, cache)
	case "local":
		return NewLocalFetcher(uri, cache)
	default:
		return nil
	}
}
