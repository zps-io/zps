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
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/fezz-io/zps/zps"
	"github.com/nightlyone/lockfile"
)

type FileFetcher struct {
	uri *url.URL

	cache    *Cache
	security Security
}

func NewFileFetcher(uri *url.URL, cache *Cache, security Security) *FileFetcher {
	return &FileFetcher{uri, cache, security}
}

func (f *FileFetcher) Refresh() error {
	configfile := filepath.Join(f.uri.Path, "config.db")

	if _, err := os.Stat(configfile); os.IsNotExist(err) {
		return errors.New("repo config not found")
	}

	s, err := os.OpenFile(configfile, os.O_RDWR, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(f.cache.GetConfig(f.uri.String()))
	if err != nil {
		return err
	}

	defer d.Close()
	if _, err := io.Copy(d, s); err != nil {
		return err
	}

	for _, osarch := range zps.Platforms() {
		err := f.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}
	packagefile := pkg.FileName()
	repofile := filepath.Join(f.uri.Path, osarch.String(), packagefile)
	cachefile := f.cache.GetFile(packagefile)

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	s, err := os.OpenFile(repofile, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	if !f.cache.Exists(cachefile) {
		d, err := os.Create(cachefile)
		if err != nil {
			return err
		}

		if _, err := io.Copy(d, s); err != nil {
			d.Close()
			return err
		}

		return d.Close()
	}

	return nil
}

func (f *FileFetcher) refresh(osarch *zps.OsArch) error {
	var err error

	metadatafile := filepath.Join(f.uri.Path, osarch.String(), "metadata.db")

	if _, err = os.Stat(filepath.Join(f.uri.Path, osarch.String())); os.IsNotExist(err) {
		return nil
	}

	if _, err = os.Stat(metadatafile); os.IsNotExist(err) {
		return nil
	}

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	if err == nil {
		s, err := os.OpenFile(metadatafile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer s.Close()

		d, err := os.Create(f.cache.GetMeta(osarch.String(), f.uri.String()))
		if err != nil {
			return err
		}

		if _, err := io.Copy(d, s); err != nil {
			d.Close()
			return err
		}

		return d.Close()
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
