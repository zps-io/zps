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
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/chuckpreslar/emission"

	"github.com/fezz-io/zps/sec"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
)

type FilePublisher struct {
	*emission.Emitter

	security Security

	uri  *url.URL
	name string

	prune int

	lockUri *url.URL
}

func NewFilePublisher(emitter *emission.Emitter, security Security, uri *url.URL, name string, prune int, lockUri *url.URL) *FilePublisher {
	return &FilePublisher{emitter, security, uri, name, prune, lockUri}
}

func (f *FilePublisher) Init() error {
	os.MkdirAll(f.uri.Path, os.FileMode(0750))

	for _, osarch := range zps.Platforms() {
		os.RemoveAll(filepath.Join(f.uri.Path, osarch.String()))
	}

	configPath := filepath.Join(f.uri.Path, "config.db")
	sigPath := filepath.Join(f.uri.Path, "config.sig")
	config := NewConfig(configPath)

	os.Remove(sigPath)

	err := config.Set("name", f.name)
	if err != nil {
		return err
	}

	keyPair, err := f.security.KeyPair(PublisherFromUri(f.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		f.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(f.uri)))
	} else {
		rsaKey, err := keyPair.RSAKey()
		if err != nil {
			return err
		}

		err = sec.SecuritySignFile(configPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
		if err != nil {
			return err
		}
	}

	return err
}

func (f *FilePublisher) Update() error {
	configPath := filepath.Join(f.uri.Path, "config.db")
	sigPath := filepath.Join(f.uri.Path, "config.sig")
	config := NewConfig(configPath)

	err := config.Set("name", f.name)
	if err != nil {
		return err
	}

	keyPair, err := f.security.KeyPair(PublisherFromUri(f.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		f.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(f.uri)))
	} else {
		rsaKey, err := keyPair.RSAKey()
		if err != nil {
			return err
		}

		err = sec.SecuritySignFile(configPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
		if err != nil {
			return err
		}
	}

	return err
}

func (f *FilePublisher) Channel(pkg string, channel string) error {
	keyPair, err := f.security.KeyPair(PublisherFromUri(f.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		f.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(f.uri)))
	}

	for _, osarch := range zps.Platforms() {
		err := f.channel(osarch, pkg, channel, keyPair)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FilePublisher) Publish(pkgs ...string) error {
	zpkgs := make(map[string]*zps.Pkg)
	for _, file := range pkgs {
		reader := zpkg.NewReader(file, "")

		err := reader.Read()
		if err != nil {
			return err
		}

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		zpkgs[file] = pkg
	}

	keyPair, err := f.security.KeyPair(PublisherFromUri(f.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		f.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(f.uri)))
	}

	for _, osarch := range zps.Platforms() {
		pkgFiles, pkgs := FilterPackagesByArch(osarch, zpkgs)

		if len(pkgFiles) > 0 {
			err := f.publish(osarch, pkgFiles, pkgs, keyPair)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FilePublisher) channel(osarch *zps.OsArch, pkg string, channel string, keyPair *KeyPairEntry) error {
	var err error

	metaPath := filepath.Join(f.uri.Path, osarch.String(), "metadata.db")
	sigPath := filepath.Join(f.uri.Path, osarch.String(), "metadata.sig")

	os.Mkdir(filepath.Join(f.uri.Path, osarch.String()), 0750)
	os.Remove(sigPath)

	locker := NewLocker(f.lockUri)

	err = locker.Lock()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", f.name, err.Error())
	}

	defer locker.Unlock()

	metadata := NewMetadata(metaPath)
	meta, err := metadata.All()
	if err != nil {
		return err
	}

	if len(meta) > 0 {
		err = metadata.Channels.Add(pkg, channel)
		if err != nil {
			return err
		}

		if keyPair != nil {
			rsaKey, err := keyPair.RSAKey()
			if err != nil {
				return err
			}

			err = sec.SecuritySignFile(metaPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
			if err != nil {
				return err
			}
		}

		f.Emit("publisher.channel", pkg)
	}

	return nil
}

func (f *FilePublisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg, keyPair *KeyPairEntry) error {
	var err error

	metaPath := filepath.Join(f.uri.Path, osarch.String(), "metadata.db")
	sigPath := filepath.Join(f.uri.Path, osarch.String(), "metadata.sig")
	repo := &zps.Repo{}

	os.Mkdir(filepath.Join(f.uri.Path, osarch.String()), 0750)
	os.Remove(sigPath)

	locker := NewLocker(f.lockUri)

	err = locker.Lock()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", f.name, err.Error())
	}

	defer locker.Unlock()

	metadata := NewMetadata(metaPath)

	meta, err := metadata.All()
	if err != nil {
		return err
	}
	repo.Load(meta)

	rejects := repo.Add(zpkgs...)
	rejectIndex := make(map[string]bool)

	for _, r := range rejects {
		rejectIndex[r.FileName()] = true
	}

	rmFiles, err := repo.Prune(f.prune)
	if err != nil {
		return err
	}

	for _, r := range rmFiles {
		rejectIndex[r.FileName()] = true
	}

	if len(repo.Solvables()) > 0 {
		for _, file := range pkgFiles {
			if !rejectIndex[filepath.Base(file)] {
				f.Emit("publisher.publish", file)
				err = f.upload(file, filepath.Join(f.uri.Path, osarch.String(), filepath.Base(file)))
				if err != nil {
					return err
				}
			}
		}

		for _, pkg := range rmFiles {
			os.Remove(filepath.Join(f.uri.Path, osarch.String(), pkg.FileName()))
		}

		// TODO Rewrite instead of atomic update for now

		metadata.Empty()

		for _, pkg := range repo.Solvables() {
			metadata.Put(pkg.(*zps.Pkg))
		}

		if keyPair != nil {
			rsaKey, err := keyPair.RSAKey()
			if err != nil {
				return err
			}

			err = sec.SecuritySignFile(metaPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
			if err != nil {
				return err
			}
		}
	} else {
		os.RemoveAll(filepath.Join(f.uri.Path, osarch.String()))
	}

	return nil
}

func (f *FilePublisher) upload(file string, dest string) error {
	s, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	return d.Close()
}
