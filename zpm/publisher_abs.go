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
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/chuckpreslar/emission"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/blobs"
	"github.com/tombuildsstuff/giovanni/storage/2019-12-12/blob/containers"

	"github.com/fezz-io/zps/cloud"
	"github.com/fezz-io/zps/sec"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
)

type ABSPublisher struct {
	*emission.Emitter

	security Security

	workPath string

	uri   *url.URL
	name  string
	prune int

	account   string
	container string
	path      string

	blobClient      *blobs.Client
	containerClient *containers.Client
}

func NewABSPublisher(emitter *emission.Emitter, security Security, workPath string, uri *url.URL, name string, prune int) *ABSPublisher {
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

	return &ABSPublisher{
		emitter,
		security,
		workPath,
		uri,
		name,
		prune,
		cloud.AzureStorageAccountFromURL(uri),
		cloud.AzureBlobContainerFromURL(uri),
		cloud.AzureBlobObjectPrefixFromURL(uri),
		&blob,
		&container,
	}
}

func (a *ABSPublisher) Init() error {
	_, err := a.containerClient.GetProperties(context.Background(), a.account, a.container)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			_, err = a.containerClient.Create(context.Background(), a.account, a.container, containers.CreateInput{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		objects, err := a.containerClient.ListBlobs(context.Background(), a.account, a.container, containers.ListBlobsInput{
			Prefix: &a.path,
		})
		if err != nil {
			return err
		}

		for _, obj := range objects.Blobs.Blobs {
			_, err = a.blobClient.Delete(context.Background(), a.account, a.container, obj.Name, blobs.DeleteInput{
				DeleteSnapshots: true,
			})
			if err != nil {
				return err
			}
		}
	}

	// Create the config db
	tmpDir, err := ioutil.TempDir(a.workPath, "init")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.db")
	sigPath := filepath.Join(tmpDir, "config.sig")
	config := NewConfig(configPath)

	err = config.Set("name", a.name)
	if err != nil {
		return err
	}

	// Upload the config db
	configDb, err := os.Open(configPath)
	if err != nil {
		return err
	}

	err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, "config.db"), configDb, blobs.PutBlockBlobInput{})
	if err != nil {
		return err
	}

	// Sign and upload
	keyPair, err := a.security.KeyPair(PublisherFromUri(a.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		a.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(a.uri)))
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

		err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, "config.sig"), configSig, blobs.PutBlockBlobInput{})
		if err != nil {
			return err
		}
	}

	return err
}

func (a *ABSPublisher) Update() error {
	tmpDir, err := ioutil.TempDir(a.workPath, "update")
	if err != nil {
		return err
	}

	configPath := filepath.Join(tmpDir, "config.db")
	sigPath := filepath.Join(tmpDir, "config.sig")

	defer os.RemoveAll(tmpDir)

	// Download the config db
	err = a.chunkedGet(path.Join(a.path, "config.db"), configPath)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", a.uri.Path))
	}

	// Modify config db
	config := NewConfig(configPath)

	err = config.Set("name", a.name)
	if err != nil {
		return err
	}

	// Upload config db
	configDb, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer configDb.Close()

	err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, "config.db"), configDb, blobs.PutBlockBlobInput{})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to upload: %s", a.uri.Path))
	}

	// Sign and upload
	keyPair, err := a.security.KeyPair(PublisherFromUri(a.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		a.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(a.uri)))
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

		err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, "config.sig"), configSig, blobs.PutBlockBlobInput{})
		if err != nil {
			return errors.New(fmt.Sprintf("unable to upload: %s", a.uri.Path))
		}
	}

	return err
}

func (a *ABSPublisher) Channel(pkg string, channel string) error {
	keyPair, err := a.security.KeyPair(PublisherFromUri(a.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		a.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(a.uri)))
	}

	for _, osarch := range zps.Platforms() {
		err := a.channel(osarch, pkg, channel, keyPair)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ABSPublisher) Publish(pkgs ...string) error {
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

	keyPair, err := a.security.KeyPair(PublisherFromUri(a.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		a.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(a.uri)))
	}

	for _, osarch := range zps.Platforms() {
		pkgFiles, pkgs := FilterPackagesByArch(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := a.publish(osarch, pkgFiles, pkgs, keyPair)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *ABSPublisher) Lock() error {
	ctx := context.Background()

	res, err := a.blobClient.Get(ctx, a.account, a.container, path.Join(a.path, ".lock"), blobs.GetInput{})

	if err != nil {
		return err
	}

	if res.StatusCode != 404 {
		return fmt.Errorf("Lock file exists")
	}

	_, err = a.blobClient.PutBlockBlob(ctx, a.account, a.container, path.Join(a.path, ".lock"), blobs.PutBlockBlobInput{})

	return err
}

func (a *ABSPublisher) Unlock() error {
	ctx := context.Background()
	_, err := a.blobClient.Delete(ctx, a.account, a.container, path.Join(a.path, ".lock"), blobs.DeleteInput{})

	return err
}

func (a *ABSPublisher) channel(osarch *zps.OsArch, pkg string, channel string, keyPair *KeyPairEntry) error {
	tmpDir, err := ioutil.TempDir(a.workPath, "channel")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	err = a.Lock()
	if err != nil {
		return errors.New("Repository: " + a.uri.String() + " is locked by another process")
	}

	defer a.Unlock()

	// Download metadata db
	err = a.chunkedGet(path.Join(a.path, osarch.String(), "metadata.db"), metaPath)
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", a.uri.Path))
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

		a.Emit("publisher.channel", pkg)
	}

	// Upload metadata
	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, osarch.String(), "metadata.db"), metadataDb, blobs.PutBlockBlobInput{})
	if err != nil {
		return err
	}

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

		err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, osarch.String(), "metadata.sig"), metadataSig, blobs.PutBlockBlobInput{})
		if err != nil {
			return err
		}
	}

	return err
}

func (a *ABSPublisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg, keyPair *KeyPairEntry) error {
	tmpDir, err := ioutil.TempDir(a.workPath, "publish")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	err = a.Lock()
	if err != nil {
		return errors.New("Repository: " + a.uri.String() + " is locked by another process")
	}

	defer a.Unlock()

	// Download metadata db
	err = a.chunkedGet(path.Join(a.path, osarch.String(), "metadata.db"), metaPath)

	if err != nil {
		if !strings.Contains(err.Error(), "404") {
			return errors.New(fmt.Sprintf("unable to download: %s", a.uri.Path))
		}
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

	rmFiles, err := repo.Prune(a.prune)
	if err != nil {
		return err
	}

	for _, r := range rmFiles {
		rejectIndex[r.FileName()] = true
	}

	if len(repo.Solvables()) > 0 {
		for _, file := range pkgFiles {
			if !rejectIndex[filepath.Base(file)] {
				a.Emit("spin.start", fmt.Sprintf("publishing: %s", file))

				err = a.upload(file, path.Join(a.path, osarch.String(), filepath.Base(file)))
				if err != nil {
					a.Emit("spin.error", fmt.Sprintf("failed: %s", file))
					return err
				}

				a.Emit("spin.success", fmt.Sprintf("published: %s", file))
			}
		}

		for _, pkg := range rmFiles {
			_, err = a.blobClient.Delete(context.Background(), a.account, a.container, path.Join(a.path, osarch.String(), pkg.FileName()), blobs.DeleteInput{
				DeleteSnapshots: true,
			})
			if err != nil {
				return err
			}
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

		err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, osarch.String(), "metadata.db"), metadataUp, blobs.PutBlockBlobInput{})
		if err != nil {
			return err
		}

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

			err = a.blobClient.PutBlockBlobFromFile(context.Background(), a.account, a.container, path.Join(a.path, osarch.String(), "metadata.sig"), metadataSig, blobs.PutBlockBlobInput{})
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (a *ABSPublisher) upload(file string, dest string) error {
	src, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer src.Close()

	ctx := context.Background()

	return a.blobClient.PutBlockBlobFromFile(ctx, cloud.AzureStorageAccountFromURL(a.uri), cloud.AzureBlobContainerFromURL(a.uri), dest, src, blobs.PutBlockBlobInput{})
}

func (a *ABSPublisher) chunkedGet(source, dest string) error {
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
