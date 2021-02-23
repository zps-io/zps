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
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/chuckpreslar/emission"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/blobs"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/containers"

	"github.com/fezz-io/zps/cloud"
	"github.com/fezz-io/zps/zps"
)

type ABSFetcher struct {
	uri *url.URL

	cache    *Cache
	security Security

	account         string
	container       string
	path            string
	blobClient      *blobs.Client
	containerClient *containers.Client
}

func NewABSFetcher(uri *url.URL, cache *Cache, security Security) *ABSFetcher {
	authorizer, err := auth.NewAuthorizerFromEnvironmentWithResource("https://storage.azure.com/")
	if err != nil {
		authorizer, err = auth.NewAuthorizerFromCLIWithResource("https://storage.azure.com/")
		if err != nil {
			return nil
		}
	}

	blob := blobs.New()
	blob.Client.Authorizer = authorizer

	container := containers.New()
	container.Client.Authorizer = authorizer

	return &ABSFetcher{
		uri,
		cache,
		security,
		cloud.AzureStorageAccountFromURL(uri),
		cloud.AzureBlobContainerFromURL(uri),
		cloud.AzureBlobObjectPrefixFromURL(uri),
		&blob,
		&container,
	}
}

func (a *ABSFetcher) Refresh() error {
	dst := a.cache.GetConfig(a.uri.String())

	err := a.chunkedGet(path.Join(a.path, "config.db"), dst)
	if err != nil {
		os.Remove(dst)

		return errors.New(fmt.Sprintf("refresh failed: %s", a.uri.String()))
	}

	if a.security.Mode() != SecurityModeNone {
		sdst := a.cache.GetConfigSig(a.uri.String())

		err = a.chunkedGet(path.Join(a.path, "config.sig"), sdst)
		if err != nil {
			os.Remove(sdst)

			return errors.New(fmt.Sprintf("refresh failed: %s", a.uri.String()))
		}

		// Validate config signature
		err = ValidateFileSignature(a.security, a.cache.GetConfig(a.uri.String()), a.cache.GetConfigSig(a.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(a.cache.GetConfig(a.uri.String()))
			os.Remove(a.cache.GetConfigSig(a.uri.String()))

			return err
		}
	}

	for _, osarch := range zps.Platforms() {
		err := a.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ABSFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}
	target := path.Join(a.path, osarch.String(), pkg.FileName())
	cacheFile := a.cache.GetFile(pkg.FileName())

	// Fetch package if not in cache
	if !a.cache.Exists(cacheFile) {
		err = a.chunkedGet(target, cacheFile)
		if err != nil {
			return errors.New(fmt.Sprintf("unable to download: %s", target))
		}
	}

	// Validate pkg
	if a.security.Mode() != SecurityModeNone {
		err = ValidateZpkg(&emission.Emitter{}, a.security, cacheFile, true)
		if err != nil {
			os.Remove(cacheFile)

			return errors.New(fmt.Sprintf("failed to validate signature: %s", pkg.FileName()))
		}
	}

	return err
}

func (a *ABSFetcher) Keys() ([][]string, error) {
	var prefix *string
	if a.path != "" {
		prefix = &a.path
	}

	objects, err := a.containerClient.ListBlobs(context.Background(), a.account, a.container, containers.ListBlobsInput{
		Prefix: prefix,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list items in bucket %q, %v", a.uri.Host, err)
	}

	var certs [][]string

	for _, item := range objects.Blobs.Blobs {
		if !strings.Contains(item.Name, ".pem") {
			continue
		}

		object, err := a.blobClient.Get(context.Background(), a.account, a.container, item.Name, blobs.GetInput{})
		if err != nil {
			return nil, fmt.Errorf("unable to fetch pem %q, %v", a.uri.Host, err)
		}

		subject, publisher, err := a.security.Trust(&object.Contents, "")
		if err != nil {
			return nil, err
		}

		certs = append(certs, []string{subject, publisher})
	}

	return certs, nil
}

func (a *ABSFetcher) refresh(osarch *zps.OsArch) error {
	var err error
	target := path.Join(a.path, osarch.String(), "metadata.db")
	dst := a.cache.GetMeta(osarch.String(), a.uri.String())

	err = a.chunkedGet(target, dst)
	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return errors.New(fmt.Sprintf("unable to download: %s", target))
		} else {
			os.Remove(dst)
			return nil
		}
	}

	if a.security.Mode() != SecurityModeNone {
		starget := path.Join(a.path, osarch.String(), "metadata.sig")
		sdst := a.cache.GetMetaSig(osarch.String(), a.uri.String())

		err = a.chunkedGet(starget, sdst)
		if err != nil {
			if !strings.Contains(err.Error(), "404") {
				return errors.New(fmt.Sprintf("unable to download: %s", target))
			} else {
				os.Remove(sdst)
				return nil
			}
		}

		// Validate metadata signature
		err = ValidateFileSignature(a.security, a.cache.GetMeta(osarch.String(), a.uri.String()), a.cache.GetMetaSig(osarch.String(), a.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(a.cache.GetMeta(osarch.String(), a.uri.String()))
			os.Remove(a.cache.GetMetaSig(osarch.String(), a.uri.String()))

			return err
		}
	}

	return nil
}

func (a *ABSFetcher) chunkedGet(source, dest string) error {
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	props, err := a.blobClient.GetProperties(context.Background(), a.account, a.container, source, blobs.GetPropertiesInput{})
	if err != nil {
		return err
	}

	offset := int64(0)
	for offset <= props.ContentLength {
		end := offset + 4*1024*1024
		if end > props.ContentLength {
			end = props.ContentLength
		}

		object, err := a.blobClient.Get(context.Background(), a.account, a.container, source, blobs.GetInput{
			StartByte: &offset,
			EndByte:   &end,
		})
		if err != nil {
			return err
		}

		_, err = dst.Write(object.Contents)
		if err != nil {
			return err
		}

		offset = end + 1
	}

	return nil
}
