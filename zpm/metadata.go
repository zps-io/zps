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
	"github.com/asdine/storm"
	"github.com/coreos/bbolt"
	"github.com/fezz-io/zps/zps"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Metadata struct {
	Path         string
	Packages     *MetadataPackages
	Channels     *MetadataChannels
	Updates      *MetadataUpdates
}

type MetadataPackages struct {
	getDb func() (*storm.DB, error)
}

type MetadataChannels struct {
	getDb func() (*storm.DB, error)
}

type MetadataUpdates struct {
	getDb func() (*storm.DB, error)
}

func NewMetadata(path string) *Metadata {
	meta := &Metadata{Path: path}

	meta.Packages = &MetadataPackages{}
	meta.Packages.getDb = meta.getDb

	meta.Channels = &MetadataChannels{}
	meta.Channels.getDb = meta.getDb

	return meta
}

func (m *Metadata) getDb() (*storm.DB, error) {
	db, err := storm.Open(filepath.Join(m.Path, "metadata.db"), storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (m *Metadata) All() ([]*zps.Pkg, error) {
	db, err := m.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*zps.PkgEntry
	var packages []*zps.Pkg

	err = db.All(&entries)

	for _, entry := range entries {
		packages = append(packages, entry.ToPkg())
	}

	return packages, nil
}

func (m *Metadata) Empty() error {
	return os.RemoveAll(filepath.Join(m.Path, "metadata.db"))
}

func (m *Metadata) Get(name string) ([]*zps.Pkg, error) {
	db, err := m.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*zps.PkgEntry
	var packages []*zps.Pkg

	err = db.Prefix("Id", name+"@", &entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	for _, entry := range entries {
		packages = append(packages, entry.ToPkg())
	}

	return packages, nil
}

func (m *Metadata) Del(id string) error {
	db, err := m.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.DeleteStruct(&zps.PkgEntry{Id: id})

	return err
}

func (m *Metadata) Put(pkg *zps.Pkg) error {
	db, err := m.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Save(pkg.ToEntry())
	return err
}

func (m *MetadataChannels) Add(id string, channel string) error {
	db, err := m.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	var entry *zps.PkgEntry

	err = db.One("Id", id, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil
		}

		return err
	}

	for _, ch := range entry.Channels {
		if ch == channel {
			return nil
		}
	}

	entry.Channels = append(entry.Channels, channel)

	err = db.Save(entry)
	return err
}

func (m *MetadataChannels) Remove(id string, channel string) error {
	db, err := m.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	var entry *zps.PkgEntry

	err = db.One("Id", id, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil
		}

		return err
	}

	var channels []string

	for _, ch := range entry.Channels {
		if ch != channel {
			channels = append(channels, ch)
		}
	}

	entry.Channels = channels

	err = db.Save(entry)
	return err
}

func (m *MetadataChannels) List() ([]string, error) {
	db, err := m.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*zps.PkgEntry

	err = db.All(&entries)

	cmap := make(map[string]bool)
	var channels []string

	for _, entry := range entries {
		for _, ch := range entry.Channels {
			cmap[ch] = true
		}
	}

	for key := range cmap {
		channels = append(channels, key)
	}

	sort.Strings(channels)

	return channels, nil
}