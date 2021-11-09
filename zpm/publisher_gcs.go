/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2021 Zachary Schneider
 */

package zpm

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/chuckpreslar/emission"
	"google.golang.org/api/iterator"

	"github.com/fezz-io/zps/sec"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
)

type GCSPublisher struct {
	*emission.Emitter

	security Security

	workPath string

	uri   *url.URL
	name  string
	prune int

	lockUri *url.URL
}

func NewGCSPublisher(emitter *emission.Emitter, security Security, workPath string, uri *url.URL, name string, prune int, lockUri *url.URL) *GCSPublisher {
	return &GCSPublisher{emitter, security, workPath, uri, name, prune, lockUri}
}

func (g *GCSPublisher) Init() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Empty the repo at prefix
	it := client.Bucket(g.uri.Host).Objects(ctx, &storage.Query{
		Prefix: g.uri.Path + "/",
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Bucket(%q).Objects(): %v", g.uri.Host, err)
		}

		delCtx, cancel := context.WithTimeout(ctx, time.Second*10)

		o := client.Bucket(g.uri.Host).Object(attrs.Name)
		if err := o.Delete(delCtx); err != nil {
			cancel()
			return fmt.Errorf("Object(%q).Delete: %v", attrs.Name, err)
		}

		cancel()
	}

	// Create repo root path
	createCtx, cancel := context.WithTimeout(ctx, time.Second*10)

	o := client.Bucket(g.uri.Host).Object(g.uri.Path + "/")
	if _, err := o.NewWriter(createCtx).Write([]byte("")); err != nil {
		cancel()
		return fmt.Errorf("Object(%q).Create: %v", g.uri.Path+"/", err)
	}
	cancel()

	// Create the config db
	tmpDir, err := ioutil.TempDir(g.workPath, "init")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.db")
	sigPath := filepath.Join(tmpDir, "config.sig")
	config := NewConfig(configPath)

	err = config.Set("name", g.name)
	if err != nil {
		return err
	}

	// Upload the config db
	configDb, err := os.Open(configPath)
	if err != nil {
		return err
	}

	cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

	cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.db")).NewWriter(cuCtx)
	if _, err = io.Copy(cu, configDb); err != nil {
		cancel()
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := cu.Close(); err != nil {
		cancel()
		return fmt.Errorf("Writer.Close: %v", err)
	}
	cancel()

	// Sign and upload
	keyPair, err := g.security.KeyPair(PublisherFromUri(g.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		g.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(g.uri)))
	} else {
		rsaKey, err := keyPair.RSAKey()
		if err != nil {
			return err
		}

		err = sec.SecuritySignFile(configPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
		if err != nil {
			return err
		}

		configSig, err := os.Open(sigPath)
		if err != nil {
			return err
		}

		cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.sig")).NewWriter(cuCtx)
		if _, err = io.Copy(cu, configSig); err != nil {
			cancel()
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := cu.Close(); err != nil {
			cancel()
			return fmt.Errorf("Writer.Close: %v", err)
		}
		cancel()
	}

	return nil
}

func (g *GCSPublisher) Update() error {
	tmpDir, err := ioutil.TempDir(g.workPath, "update")
	if err != nil {
		return err
	}

	configPath := filepath.Join(tmpDir, "config.db")
	sigPath := filepath.Join(tmpDir, "config.sig")

	defer os.RemoveAll(tmpDir)

	configDb, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configDb.Close()

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
		return fmt.Errorf("unable to download: %s", g.uri.Path)
	}
	if _, err = io.Copy(configDb, cd); err != nil {
		cancel()
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := cd.Close(); err != nil {
		cancel()
		return fmt.Errorf("Reader.Close: %v", err)
	}
	cancel()

	// Modify config db
	config := NewConfig(configPath)

	err = config.Set("name", g.name)
	if err != nil {
		return err
	}

	// Upload config db
	configDb.Seek(0, io.SeekStart)

	cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

	cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.db")).NewWriter(cuCtx)
	if _, err = io.Copy(cu, configDb); err != nil {
		cancel()
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := cu.Close(); err != nil {
		cancel()
		return fmt.Errorf("Writer.Close: %v", err)
	}
	cancel()

	// Sign and upload
	keyPair, err := g.security.KeyPair(PublisherFromUri(g.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		g.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(g.uri)))
	} else {
		rsaKey, err := keyPair.RSAKey()
		if err != nil {
			return err
		}

		err = sec.SecuritySignFile(configPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
		if err != nil {
			return err
		}

		configSig, err := os.Open(sigPath)
		if err != nil {
			return err
		}

		cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, "config.sig")).NewWriter(cuCtx)
		if _, err = io.Copy(cu, configSig); err != nil {
			cancel()
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := cu.Close(); err != nil {
			cancel()
			return fmt.Errorf("Writer.Close: %v", err)
		}
		cancel()
	}

	return err
}

func (g *GCSPublisher) Channel(pkg string, channel string) error {
	keyPair, err := g.security.KeyPair(PublisherFromUri(g.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		g.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(g.uri)))
	}

	for _, osarch := range zps.Platforms() {
		err := g.channel(osarch, pkg, channel, keyPair)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *GCSPublisher) Publish(pkgs ...string) error {
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

	keyPair, err := g.security.KeyPair(PublisherFromUri(g.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		g.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(g.uri)))
	}

	for _, osarch := range zps.Platforms() {
		pkgFiles, pkgs := FilterPackagesByArch(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := g.publish(osarch, pkgFiles, pkgs, keyPair)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GCSPublisher) channel(osarch *zps.OsArch, pkg string, channel string, keyPair *KeyPairEntry) error {
	retries := 5

	var newEtag [16]byte

	tmpDir, err := ioutil.TempDir(g.workPath, "channel")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	locker := NewLocker(g.lockUri)

	eTag, err := locker.LockWithEtag()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", g.name, err.Error())
	}

	defer locker.UnlockWithEtag(newEtag)

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	for {
		// Download metadata db
		cdCtx, cancel := context.WithTimeout(ctx, time.Second*60)
		cd, err := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.db")).NewReader(cdCtx)
		defer cancel()
		if err != nil {
			cancel()
			return fmt.Errorf("unable to download: %s", g.uri.Path)
		}

		if _, err = io.Copy(metadataDb, cd); err != nil {
			cancel()
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := cd.Close(); err != nil {
			cancel()
			return fmt.Errorf("Reader.Close: %v", err)
		}

		medatabaDbData := make([]byte, 0)

		_, err = metadataDb.Read(medatabaDbData)
		if err != nil {
			return fmt.Errorf("unable to read downloaded file: %s", g.uri.Path)
		}

		actualETag := md5.Sum(medatabaDbData)

		// if eTag is empty, it means locker method doesn't support
		// storing attribute or doesn't contain previous eTag
		// we will break from cycle right after
		if len(eTag) > 0 && actualETag != eTag {
			retries -= 1
			if retries == 0 {
				return fmt.Errorf("S3 object %q has eTag mismatch: want %q, got %q", g.uri.Path, eTag, actualETag)
			}
		}

		break

	}

	// Modify metadata
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

		g.Emit("publisher.channel", pkg)
	}

	// Upload metadata
	metadataDb.Seek(0, io.SeekStart)

	cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

	cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.db")).NewWriter(cuCtx)
	if _, err = io.Copy(cu, metadataDb); err != nil {
		cancel()
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := cu.Close(); err != nil {
		cancel()
		return fmt.Errorf("Writer.Close: %v", err)
	}
	cancel()

	medatabaDbData := make([]byte, 0)
	_, err = metadataDb.Read(medatabaDbData)
	if err != nil {
		return fmt.Errorf("unable to read new metadata file: %s", g.uri.Path)
	}

	eTag = md5.Sum(medatabaDbData)

	// Sign and upload
	if keyPair != nil {
		rsaKey, err := keyPair.RSAKey()
		if err != nil {
			return err
		}

		err = sec.SecuritySignFile(metaPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
		if err != nil {
			return err
		}

		metadataSig, err := os.Open(sigPath)
		if err != nil {
			return err
		}

		cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.sig")).NewWriter(cuCtx)
		if _, err = io.Copy(cu, metadataSig); err != nil {
			cancel()
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := cu.Close(); err != nil {
			cancel()
			return fmt.Errorf("Writer.Close: %v", err)
		}
		cancel()
	}

	return err
}

func (g *GCSPublisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg, keyPair *KeyPairEntry) error {
	retries := 5

	var newEtag [16]byte

	tmpDir, err := ioutil.TempDir(g.workPath, "publish")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	locker := NewLocker(g.lockUri)

	eTag, err := locker.LockWithEtag()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", g.name, err.Error())
	}

	defer locker.UnlockWithEtag(newEtag)

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	for {
		// Download metadata db
		cdCtx, cancel := context.WithTimeout(ctx, time.Second*60)
		defer cancel()

		cd, err := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.db")).NewReader(cdCtx)
		if cd != nil {
			if _, err = io.Copy(metadataDb, cd); err != nil {
				cancel()
				return fmt.Errorf("io.Copy: %v", err)
			}
			if err := cd.Close(); err != nil {
				cancel()
				return fmt.Errorf("Reader.Close: %v", err)
			}
		}

		medatabaDbData := make([]byte, 0)

		_, err = metadataDb.Read(medatabaDbData)
		if err != nil {
			return fmt.Errorf("unable to read downloaded file: %s", g.uri.Path)
		}

		actualETag := md5.Sum(medatabaDbData)

		// if eTag is empty, it means locker method doesn't support
		// storing attribute or doesn't contain previous eTag
		// we will break from cycle right after
		if len(eTag) > 0 && actualETag != eTag {
			retries -= 1
			if retries == 0 {
				return fmt.Errorf("S3 object %q has eTag mismatch: want %q, got %q", g.uri.Path, eTag, actualETag)
			}
		}

		break
	}

	metadata := NewMetadata(metaPath)
	repo := &zps.Repo{}

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

	rmFiles, err := repo.Prune(g.prune)
	if err != nil {
		return err
	}

	for _, r := range rmFiles {
		rejectIndex[r.FileName()] = true
	}

	if len(repo.Solvables()) > 0 {
		for _, file := range pkgFiles {
			if !rejectIndex[filepath.Base(file)] {
				g.Emit("spin.start", fmt.Sprintf("publishing: %s", file))

				err = g.upload(file, path.Join(g.uri.Path, osarch.String(), filepath.Base(file)))
				if err != nil {
					g.Emit("spin.error", fmt.Sprintf("failed: %s", file))
					return err
				}

				g.Emit("spin.success", fmt.Sprintf("published: %s", file))
			}
		}

		for _, pkg := range rmFiles {
			delCtx, cancel := context.WithTimeout(ctx, time.Second*10)

			o := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), pkg.FileName()))
			if err := o.Delete(delCtx); err != nil {
				cancel()
				return fmt.Errorf("Object(%q).Delete: %v", path.Join(g.uri.Path, osarch.String(), pkg.FileName()), err)
			}
			cancel()
		}

		// TODO Rewrite instead of atomic update for now

		metadata.Empty()

		for _, pkg := range repo.Solvables() {
			err := metadata.Put(pkg.(*zps.Pkg))
			if err != nil {
				return err
			}
		}

		// Upload metadata
		metadataUp, err := os.Open(metaPath)
		if err != nil {
			return err
		}
		defer metadataUp.Close()

		cuCtx, cancel := context.WithTimeout(ctx, time.Second*60)

		cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.db")).NewWriter(cuCtx)
		if _, err = io.Copy(cu, metadataUp); err != nil {
			cancel()
			return fmt.Errorf("io.Copy: %v", err)
		}
		if err := cu.Close(); err != nil {
			cancel()
			return fmt.Errorf("Writer.Close: %v", err)
		}
		cancel()

		// Sign and upload
		if keyPair != nil {
			rsaKey, err := keyPair.RSAKey()
			if err != nil {
				return err
			}

			err = sec.SecuritySignFile(metaPath, sigPath, keyPair.Fingerprint, rsaKey, sec.DefaultDigestMethod)
			if err != nil {
				return err
			}

			metadataSig, err := os.Open(sigPath)
			if err != nil {
				return err
			}

			cuCtx, cancel := context.WithTimeout(ctx, time.Second*50)

			cu := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String(), "metadata.sig")).NewWriter(cuCtx)
			if _, err = io.Copy(cu, metadataSig); err != nil {
				cancel()
				return fmt.Errorf("io.Copy: %v", err)
			}
			if err := cu.Close(); err != nil {
				cancel()
				return fmt.Errorf("Writer.Close: %v", err)
			}
			cancel()
		}
	} else {
		delCtx, cancel := context.WithTimeout(ctx, time.Second*10)

		o := client.Bucket(g.uri.Host).Object(path.Join(g.uri.Path, osarch.String()) + "/")
		if err := o.Delete(delCtx); err != nil {
			cancel()
			return fmt.Errorf("Object(%q).Delete: %v", path.Join(g.uri.Path, osarch.String())+"/", err)
		}
		cancel()
	}

	return err
}

func (g *GCSPublisher) upload(file string, dest string) error {
	src, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer src.Close()

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	cuCtx, cancel := context.WithTimeout(ctx, time.Second*900)
	defer cancel()

	cu := client.Bucket(g.uri.Host).Object(dest).NewWriter(cuCtx)
	if _, err = io.Copy(cu, src); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := cu.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return err
}
