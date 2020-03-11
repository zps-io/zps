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
	"net/url"
	"os"
	"path"
	"time"

	"github.com/chuckpreslar/emission"

	"github.com/fezz-io/zps/zps"
	"gopkg.in/resty.v1"
)

type HttpsFetcher struct {
	uri *url.URL

	cache    *Cache
	security Security

	client *resty.Client
}

func NewHttpsFetcher(uri *url.URL, cache *Cache, security Security) *HttpsFetcher {
	client := resty.New()
	client.SetTimeout(time.Duration(10) * time.Second)

	return &HttpsFetcher{uri, cache, security, client}
}

func (h *HttpsFetcher) Refresh() error {
	configUri, _ := url.Parse(h.uri.String())
	configUri.Path = path.Join(configUri.Path, "config.db")

	user := configUri.User.Username()
	password, _ := configUri.User.Password()

	resp, err := h.client.R().
		SetBasicAuth(user, password).
		SetOutput(h.cache.GetConfig(h.uri.String())).
		Get(configUri.String())

	if err != nil {
		return errors.New(fmt.Sprintf("error connecting to: %s", h.uri.Host))
	}

	if resp.IsError() {
		switch resp.StatusCode() {
		case 404:
			return errors.New(fmt.Sprintf("not found: %s", configUri.String()))
		case 403:
			return errors.New(fmt.Sprintf("access denied: %s", configUri.String()))
		default:
			return errors.New(fmt.Sprintf("server error %d: %s", resp.StatusCode(), configUri.String()))
		}
	}

	if h.security.Mode() != SecurityModeNone {
		sigUri, _ := url.Parse(h.uri.String())
		sigUri.Path = path.Join(sigUri.Path, "config.sig")

		resp, err := h.client.R().
			SetBasicAuth(user, password).
			SetOutput(h.cache.GetConfigSig(h.uri.String())).
			Get(sigUri.String())

		if err != nil {
			return errors.New(fmt.Sprintf("error connecting to: %s", h.uri.Host))
		}

		if resp.IsError() {
			switch resp.StatusCode() {
			case 404:
				return errors.New(fmt.Sprintf("not found: %s", sigUri.String()))
			case 403:
				return errors.New(fmt.Sprintf("access denied: %s", sigUri.String()))
			default:
				return errors.New(fmt.Sprintf("server error %d: %s", resp.StatusCode(), sigUri.String()))
			}
		}

		// Validate config signature
		err = ValidateFileSignature(h.security, h.cache.GetConfig(h.uri.String()), h.cache.GetConfigSig(h.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(h.cache.GetConfig(h.uri.String()))
			os.Remove(h.cache.GetConfigSig(h.uri.String()))

			return err
		}
	}

	for _, osarch := range zps.Platforms() {
		err := h.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HttpsFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}

	fileUri, _ := url.Parse(h.uri.String())
	fileUri.Path = path.Join(fileUri.Path, osarch.String(), pkg.FileName())

	cacheFile := h.cache.GetFile(pkg.FileName())

	user := fileUri.User.Username()
	password, _ := fileUri.User.Password()

	// Fetch package if not in cache
	if !h.cache.Exists(cacheFile) {
		resp, err := h.client.R().
			SetBasicAuth(user, password).
			SetOutput(cacheFile).
			Get(fileUri.String())

		if err != nil {
			return errors.New(fmt.Sprintf("error connecting to: %s", h.uri.Host))
		}

		if resp.IsError() {
			switch resp.StatusCode() {
			case 404:
				return errors.New(fmt.Sprintf("not found: %s", fileUri.String()))
			case 403:
				return errors.New(fmt.Sprintf("access denied: %s", fileUri.String()))
			default:
				return errors.New(fmt.Sprintf("server error %d: %s", resp.StatusCode(), fileUri.String()))
			}
		}
	}

	// Validate pkg
	if h.security.Mode() != SecurityModeNone {
		err = ValidateZpkg(&emission.Emitter{}, h.security, cacheFile, true)
		if err != nil {
			os.Remove(cacheFile)

			return errors.New(fmt.Sprintf("failed to validate signature: %s", pkg.FileName()))
		}
	}

	return nil
}

func (h *HttpsFetcher) refresh(osarch *zps.OsArch) error {
	var err error

	metadataUri, _ := url.Parse(h.uri.String())
	metadataUri.Path = path.Join(metadataUri.Path, osarch.String(), "metadata.db")

	user := metadataUri.User.Username()
	password, _ := metadataUri.User.Password()

	resp, err := h.client.R().
		SetBasicAuth(user, password).
		SetOutput(h.cache.GetMeta(osarch.String(), h.uri.String())).
		Get(metadataUri.String())

	if err != nil {
		return errors.New(fmt.Sprintf("error connecting to: %s", h.uri.Host))
	}

	if resp.IsError() {
		switch resp.StatusCode() {
		case 404:
			return nil
		case 403:
			return nil
		default:
			return errors.New(fmt.Sprintf("server error %d: %s", resp.StatusCode(), metadataUri.String()))
		}
	}

	if h.security.Mode() != SecurityModeNone {
		sigUri, _ := url.Parse(h.uri.String())
		sigUri.Path = path.Join(sigUri.Path, osarch.String(), "metadata.sig")

		resp, err := h.client.R().
			SetBasicAuth(user, password).
			SetOutput(h.cache.GetMetaSig(osarch.String(), h.uri.String())).
			Get(sigUri.String())

		if err != nil {
			return errors.New(fmt.Sprintf("error connecting to: %s", h.uri.Host))
		}

		if resp.IsError() {
			switch resp.StatusCode() {
			case 404:
				return errors.New(fmt.Sprintf("not found: %s", sigUri.String()))
			case 403:
				return errors.New(fmt.Sprintf("access denied: %s", sigUri.String()))
			default:
				return errors.New(fmt.Sprintf("server error %d: %s", resp.StatusCode(), sigUri.String()))
			}
		}

		// Validate config signature
		err = ValidateFileSignature(h.security, h.cache.GetMeta(osarch.String(), h.uri.String()), h.cache.GetMetaSig(osarch.String(), h.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(h.cache.GetMeta(osarch.String(), h.uri.String()))
			os.Remove(h.cache.GetMetaSig(osarch.String(), h.uri.String()))

			return err
		}
	}

	return nil
}
