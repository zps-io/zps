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
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/chuckpreslar/emission"

	"github.com/fezz-io/zps/zps"
)

type LocalFetcher struct {
	uri *url.URL

	cache    *Cache
	security Security
}

func NewLocalFetcher(uri *url.URL, cache *Cache, security Security) *LocalFetcher {
	return &LocalFetcher{uri, cache, security}
}

func (f *LocalFetcher) Refresh() error {
	return nil
}

func (f *LocalFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	packageFile := pkg.FileName()
	repoFile := filepath.Join(f.uri.Path, packageFile)
	cacheFile := f.cache.GetFile(packageFile)

	// Copy package if not in cache
	if !f.cache.Exists(cacheFile) {
		src, err := os.Open(repoFile)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
	}

	// Validate pkg
	if f.security.Mode() != SecurityModeNone {
		err = ValidateZpkg(&emission.Emitter{}, f.security, cacheFile, true)
		if err != nil {
			os.Remove(cacheFile)

			return errors.New(fmt.Sprintf("failed to validate signature: %s", packageFile))
		}
	}

	return nil
}

func (f *LocalFetcher) Keys() ([][]string, error) {
	return nil, errors.New("fetcher.local.keys not implemented")
}
