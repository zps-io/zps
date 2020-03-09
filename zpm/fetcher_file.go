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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/action"

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
	configFile := filepath.Join(f.uri.Path, "config.db")

	// Refresh config db
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return errors.New("repo config not found")
	}

	srcCfg, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer srcCfg.Close()

	dstCfg, err := os.OpenFile(f.cache.GetConfig(f.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dstCfg.Close()

	if _, err := io.Copy(dstCfg, srcCfg); err != nil {
		return err
	}

	if f.security.Mode() != SecurityModeNone {
		configSig := filepath.Join(f.uri.Path, "config.sig")

		// Refresh config signature
		if _, err := os.Stat(configSig); os.IsNotExist(err) {
			return errors.New("repo config signature not found")
		}

		srcSig, err := os.Open(configSig)
		if err != nil {
			return err
		}
		defer srcSig.Close()

		destSig, err := os.OpenFile(f.cache.GetConfigSig(f.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer destSig.Close()

		if _, err := io.Copy(destSig, srcSig); err != nil {
			return err
		}

		// Validate config signature
		cfgBytes, err := ioutil.ReadFile(f.cache.GetConfig(f.uri.String()))
		if err != nil {
			return err
		}

		sigBytes, err := ioutil.ReadFile(f.cache.GetConfigSig(f.uri.String()))
		if err != nil {
			return err
		}

		sig := &action.Signature{}

		err = json.Unmarshal(sigBytes, sig)
		if err != nil {
			return err
		}

		_, err = f.security.Verify(&cfgBytes, []*action.Signature{sig})
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(f.cache.GetConfig(f.uri.String()))
			os.Remove(f.cache.GetConfigSig(f.uri.String()))

			return err
		}
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
	packageFile := pkg.FileName()
	repoFile := filepath.Join(f.uri.Path, osarch.String(), packageFile)
	cacheFile := f.cache.GetFile(packageFile)

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	s, err := os.Open(repoFile)
	if err != nil {
		return err
	}
	defer s.Close()

	if !f.cache.Exists(cacheFile) {
		d, err := os.Create(cacheFile)
		if err != nil {
			return err
		}

		if _, err := io.Copy(d, s); err != nil {
			d.Close()
			return err
		}

		// Validate pkg
		if f.security.Mode() != SecurityModeNone {
			err = ValidateZpkg(&emission.Emitter{}, f.security, cacheFile, true)
			if err != nil {
				os.Remove(cacheFile)

				return errors.New(fmt.Sprintf("failed to validate signature: %s", packageFile))
			}
		}

		return d.Close()
	}

	return nil
}

func (f *FileFetcher) refresh(osarch *zps.OsArch) error {
	var err error

	metadataFile := filepath.Join(f.uri.Path, osarch.String(), "metadata.db")

	if _, err = os.Stat(filepath.Join(f.uri.Path, osarch.String())); os.IsNotExist(err) {
		return nil
	}

	if _, err = os.Stat(metadataFile); os.IsNotExist(err) {
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
		// Fetch meta
		srcMeta, err := os.Open(metadataFile)
		if err != nil {
			return err
		}
		defer srcMeta.Close()

		dstMeta, err := os.OpenFile(f.cache.GetMeta(osarch.String(), f.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer dstMeta.Close()

		if _, err := io.Copy(dstMeta, srcMeta); err != nil {
			return err
		}

		if f.security.Mode() != SecurityModeNone {
			metadataSig := filepath.Join(f.uri.Path, osarch.String(), "metadata.sig")

			// Fetch meta sig
			srcSig, err := os.Open(metadataSig)
			if err != nil {
				return err
			}
			defer srcMeta.Close()

			dstSig, err := os.OpenFile(f.cache.GetMetaSig(osarch.String(), f.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
			if err != nil {
				return err
			}
			defer dstMeta.Close()

			if _, err := io.Copy(dstSig, srcSig); err != nil {
				return err
			}

			// Validate meta signature
			metaBytes, err := ioutil.ReadFile(f.cache.GetMeta(osarch.String(), f.uri.String()))
			if err != nil {
				return err
			}

			sigBytes, err := ioutil.ReadFile(f.cache.GetMetaSig(osarch.String(), f.uri.String()))
			if err != nil {
				return err
			}

			sig := &action.Signature{}

			err = json.Unmarshal(sigBytes, sig)
			if err != nil {
				return err
			}

			_, err = f.security.Verify(&metaBytes, []*action.Signature{sig})
			if err != nil {
				// Remove the config and sig if validation fails
				os.Remove(f.cache.GetMeta(osarch.String(), f.uri.String()))
				os.Remove(f.cache.GetMetaSig(osarch.String(), f.uri.String()))

				return err
			}
		}

		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
