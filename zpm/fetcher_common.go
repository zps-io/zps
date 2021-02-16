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
	"fmt"
	"github.com/fezz-io/zps/cloud"
	"net/url"

	"github.com/fezz-io/zps/zps"
)

type Fetcher interface {
	Refresh() error
	Fetch(pkg *zps.Pkg) error
	Keys() ([][]string, error)
}

func NewFetcher(uri *url.URL, cache *Cache, security Security, cloudProvider string) Fetcher {
	switch uri.Scheme {
	case "file":
		return NewFileFetcher(uri, cache, security)
	case "https":
		return NewHttpsFetcher(uri, cache, security)
	case "local":
		return NewLocalFetcher(uri, cache, security)
	case "gcs":
		return NewGCSFetcher(uri, cache, security)
	case "s3":
		return NewS3Fetcher(uri, cache, security)
	case "cloud":
		switch cloudProvider {
		case cloud.AWS:
			return NewS3Fetcher(uri, cache, security)
		case cloud.GCP:
			return NewGCSFetcher(uri, cache, security)
		default:
			return nil
		}
	default:
		return nil
	}
}

func SafeURI(uri *url.URL) string {
	return fmt.Sprintf("%s://%s%s", uri.Scheme, uri.Host, uri.Path)
}