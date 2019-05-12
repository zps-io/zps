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
	"path"
	"time"

	"github.com/fezz-io/zps/zps"
	"gopkg.in/resty.v1"
)

type HttpsFetcher struct {
	uri    *url.URL
	cache  *Cache
	client *resty.Client
}

func NewHttpsFetcher(uri *url.URL, cache *Cache) *HttpsFetcher {
	client := resty.New()
	client.SetTimeout(time.Duration(10) * time.Second)

	return &HttpsFetcher{uri, cache, client}
}

func (h *HttpsFetcher) Refresh() error {
	configUri, _ := url.Parse(h.uri.String())
	configUri.Path = path.Join(configUri.Path, "config.db")

	resp, err := h.client.R().
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

	resp, err := h.client.R().
		SetOutput(h.cache.GetFile(pkg.FileName())).
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

	return nil
}

func (h *HttpsFetcher) refresh(osarch *zps.OsArch) error {
	var err error

	metadataUri, _ := url.Parse(h.uri.String())
	metadataUri.Path = path.Join(metadataUri.Path, osarch.String(), "metadata.db")

	resp, err := h.client.R().
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

	return nil
}
