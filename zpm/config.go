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
	"path/filepath"
	"time"
)

type Config struct {
	Path         string
}

type ConfigEntry struct {
	Key string `storm:"id"`
	Value string
}

func NewConfig(path string) *Config {
	cfg := &Config{Path: path}

	return cfg
}

func (c *Config) getDb() (*storm.DB, error) {
	db, err := storm.Open(filepath.Join(c.Path, "config.db"), storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *Config) Get(key string) (string, error) {
	db, err := c.getDb()
	if err != nil {
		return "", err
	}
	defer db.Close()

	var entry *ConfigEntry

	err = db.One("Key", key, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return "", nil
		}

		return "", err
	}

	return entry.Value, nil
}

func (m *Config) Set(key string, value string) error {
	entry := &ConfigEntry{key, value}

	db, err := m.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Save(entry)
	return err
}


