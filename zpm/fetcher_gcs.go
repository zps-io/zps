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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/chuckpreslar/emission"
	"google.golang.org/api/iterator"

	"github.com/fezz-io/zps/zps"
)

type GCSFetcher struct {
	uri *url.URL

	cache    *Cache
	security Security
}

func NewGCSFetcher(uri *url.URL, cache *Cache, security Security) *GCSFetcher {
	return &GCSFetcher{uri, cache, security}
}

func (g *GCSFetcher) Refresh() error {
	dst, err := os.OpenFile(g.cache.GetConfig(g.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Download the config db
	cdCtx, cancel := context.WithTimeout(ctx, time.Second*60)

	cd, err := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.db")).NewReader(cdCtx)
	if err != nil {
		cancel()
		dst.Close()
		os.Remove(g.cache.GetConfig(g.uri.String()))

		return fmt.Errorf("refresh failed: %s", g.uri.String())
	}
	if _, err = io.Copy(dst, cd); err != nil {
		cancel()
		dst.Close()
		os.Remove(g.cache.GetConfig(g.uri.String()))

		return fmt.Errorf("refresh failed: %s", g.uri.String())
	}
	if err := cd.Close(); err != nil {
		cancel()
		return fmt.Errorf("Reader.Close: %v", err)
	}
	cancel()

	if g.security.Mode() != SecurityModeNone {
		sdst, err := os.OpenFile(g.cache.GetConfigSig(g.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer sdst.Close()

		csCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		cs, err := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.sig")).NewReader(csCtx)
		if err != nil {
			cancel()
			sdst.Close()
			os.Remove(g.cache.GetConfigSig(g.uri.String()))

			return fmt.Errorf("refresh failed: %s", g.uri.String())
		}
		if _, err = io.Copy(sdst, cs); err != nil {
			cancel()
			sdst.Close()
			os.Remove(g.cache.GetConfigSig(g.uri.String()))

			return fmt.Errorf("refresh failed: %s", g.uri.String())
		}
		if err := cs.Close(); err != nil {
			cancel()
			return fmt.Errorf("Reader.Close: %v", err)
		}
		cancel()

		// Validate config signature
		err = ValidateFileSignature(g.security, g.cache.GetConfig(g.uri.String()), g.cache.GetConfigSig(g.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(g.cache.GetConfig(g.uri.String()))
			os.Remove(g.cache.GetConfigSig(g.uri.String()))

			return err
		}
	}

	for _, osarch := range zps.Platforms() {
		err := g.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GCSFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}
	target := path.Join(g.uri.Path, osarch.String(), pkg.FileName())
	cacheFile := g.cache.GetFile(pkg.FileName())

	// Fetch package if not in cache
	if !g.cache.Exists(cacheFile) {
		dst, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer dst.Close()

		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			return err
		}

		cdCtx, cancel := context.WithTimeout(ctx, time.Second*600)

		cd, err := client.Bucket(g.uri.Host).Object(target).NewReader(cdCtx)
		if err != nil {
			cancel()
			dst.Close()
			os.Remove(cacheFile)

			return fmt.Errorf("unable to download: %s", target)
		}
		if _, err = io.Copy(dst, cd); err != nil {
			cancel()
			dst.Close()
			os.Remove(cacheFile)

			return fmt.Errorf("unable to download: %s", target)
		}
		if err := cd.Close(); err != nil {
			cancel()
			return fmt.Errorf("Reader.Close: %v", err)
		}
		cancel()
	}

	// Validate pkg
	if g.security.Mode() != SecurityModeNone {
		err = ValidateZpkg(&emission.Emitter{}, g.security, cacheFile, true)
		if err != nil {
			os.Remove(cacheFile)

			return fmt.Errorf("failed to validate signature: %s", pkg.FileName())
		}
	}

	return err
}

func (g *GCSFetcher) Keys() ([][]string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	it := client.Bucket(g.uri.Host).Objects(ctx, &storage.Query{
		Prefix: g.uri.Path + "/",
	})

	var certs [][]string

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to list items in bucket %q, %v", g.uri.Host, err)
		}
		if !strings.Contains(attrs.Name, ".pem") {
			continue
		}

		getCtx, cancel := context.WithTimeout(ctx, time.Second*30)

		rc, err := client.Bucket(g.uri.Host).Object(attrs.Name).NewReader(getCtx)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("unable to fetch %q, %v", attrs.Name, err)
		}
		defer rc.Close()

		pem, err := ioutil.ReadAll(rc)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("unable to fetch %q, %v", attrs.Name, err)
		}
		cancel()

		subject, publisher, err := g.security.Trust(&pem, "")
		if err != nil {
			return nil, err
		}

		certs = append(certs, []string{subject, publisher})
	}

	return certs, nil
}

func (g *GCSFetcher) refresh(osarch *zps.OsArch) error {
	var err error
	target := path.Join(g.uri.Path, osarch.String(), "metadata.db")
	dest := g.cache.GetMeta(osarch.String(), g.uri.String())

	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Download metadata db
	mdCtx, cancel := context.WithTimeout(ctx, time.Second*60)
	md, err := client.Bucket(g.uri.Host).Object(target).NewReader(mdCtx)
	if err != nil {
		cancel()
		if !strings.Contains(err.Error(), "object doesn't exist") {
			return fmt.Errorf("unable to download: %s", target)
		} else {
			dst.Close()
			os.Remove(dest)
			return nil
		}
	}
	if _, err = io.Copy(dst, md); err != nil {
		cancel()
		dst.Close()
		os.Remove(dest)

		return fmt.Errorf("unable to download: %s", target)
	}
	if err := md.Close(); err != nil {
		cancel()
		return fmt.Errorf("Reader.Close: %v", err)
	}
	cancel()

	if g.security.Mode() != SecurityModeNone {
		starget := path.Join(g.uri.Path, osarch.String(), "metadata.sig")
		sdest := g.cache.GetMetaSig(osarch.String(), g.uri.String())

		sdst, err := os.OpenFile(sdest, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer sdst.Close()

		msCtx, cancel := context.WithTimeout(ctx, time.Second*60)
		ms, err := client.Bucket(g.uri.Host).Object(starget).NewReader(msCtx)
		if err != nil {
			cancel()

			if !strings.Contains(err.Error(), "object doesn't exist") {
				return fmt.Errorf("unable to download: %s", starget)
			} else {
				dst.Close()
				os.Remove(sdest)
				return nil
			}
		}
		if _, err = io.Copy(sdst, ms); err != nil {
			cancel()
			dst.Close()
			os.Remove(sdest)

			return fmt.Errorf("unable to download: %s", starget)
		}
		if err := md.Close(); err != nil {
			cancel()
			return fmt.Errorf("Reader.Close: %v", err)
		}
		cancel()

		// Validate metadata signature
		err = ValidateFileSignature(g.security, g.cache.GetMeta(osarch.String(), g.uri.String()), g.cache.GetMetaSig(osarch.String(), g.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(g.cache.GetMeta(osarch.String(), g.uri.String()))
			os.Remove(g.cache.GetMetaSig(osarch.String(), g.uri.String()))

			return err
		}
	}

	return nil
}
