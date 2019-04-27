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
	"net/url"
	"os"
	"path/filepath"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
	"github.com/nightlyone/lockfile"
	"io"
)

type FilePublisher struct {
	*emission.Emitter
	uri   *url.URL
	name  string
	prune int
}

func NewFilePublisher(emitter *emission.Emitter, uri *url.URL, name string, prune int) *FilePublisher {
	return &FilePublisher{emitter, uri, name, prune}
}

func (f *FilePublisher) Init() error {
	os.MkdirAll(f.uri.Path, os.FileMode(0750))

	for _, osarch := range zps.Platforms() {
		os.RemoveAll(filepath.Join(f.uri.Path, osarch.String()))
	}

	return f.configure()
}

func (f *FilePublisher) Update() error {
	return f.configure()
}

func (f *FilePublisher) Channel(pkg string, channel string) error {

	for _, osarch := range zps.Platforms() {

		err := f.channel(osarch, pkg, channel)
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

	for _, osarch := range zps.Platforms() {

		pkgFiles, pkgs := f.filter(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := f.publish(osarch, pkgFiles, pkgs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FilePublisher) channel(osarch *zps.OsArch, pkg string, channel string) error {
	var err error

	metapath := filepath.Join(f.uri.Path, osarch.String(), "metadata.db")

	os.Mkdir(filepath.Join(f.uri.Path, osarch.String()), 0750)

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	metadata := NewMetadata(metapath)
	meta, err := metadata.All()
	if err != nil {
		return err
	}

	if len(meta) > 0 {
		err = metadata.Channels.Add(pkg, channel)
		if err != nil {
			return err
		}

		f.Emit("publisher.channel", pkg)
	}

	return nil
}

func (f *FilePublisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg) error {
	var err error

	metapath := filepath.Join(f.uri.Path, osarch.String())
	repo := &zps.Repo{}

	os.Mkdir(filepath.Join(f.uri.Path, osarch.String()), 0750)

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	metadata := NewMetadata(metapath)

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
	} else {
		os.RemoveAll(filepath.Join(f.uri.Path, osarch.String()))
	}

	return nil
}

func (f *FilePublisher) filter(osarch *zps.OsArch, zpkgs map[string]*zps.Pkg) ([]string, []*zps.Pkg) {
	var files []string
	var pkgs []*zps.Pkg

	for k, v := range zpkgs {
		if v.Os() == osarch.Os && v.Arch() == osarch.Arch {
			files = append(files, k)
			pkgs = append(pkgs, zpkgs[k])
		}
	}

	return files, pkgs
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

// Temporary
func (f *FilePublisher) configure() error {
	config := NewConfig(f.uri.Path)

	return config.Set("name", f.name)
}
