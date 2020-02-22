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
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
)

type S3Publisher struct {
	*emission.Emitter

	workPath string

	uri   *url.URL
	name  string
	prune int

	keyPair *KeyPairEntry

	session *session.Session
}

func NewS3Publisher(emitter *emission.Emitter, workPath string, uri *url.URL, name string, prune int, keyPair *KeyPairEntry) *S3Publisher {
	sess := session.Must(session.NewSession())

	user := uri.User.Username()
	password, _ := uri.User.Password()

	if user != "" && password != "" {
		sess.Config.Credentials = credentials.NewStaticCredentials(user, password, "")
	} else {
		sess.Config.Credentials = credentials.NewEnvCredentials()
	}

	region, err := s3manager.GetBucketRegion(context.Background(), sess, uri.Host, "us-west-2")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			emitter.Emit("error", fmt.Sprintf("unable to find bucket %s's region not found\n", uri.Host))
		}

		return nil
	}

	sess.Config.Region = aws.String(region)

	return &S3Publisher{emitter, workPath, uri, name, prune, keyPair, sess}
}

func (s *S3Publisher) Init() error {
	svc := s3.New(s.session)

	input := &s3.ListObjectsInput{
		Bucket:  aws.String(s.uri.Host),
		MaxKeys: aws.Int64(100),
		Prefix:  aws.String(strings.TrimPrefix(s.uri.Path, "/") + "/"),
	}

	iter := s3manager.NewDeleteListIterator(svc, input)
	batcher := s3manager.NewBatchDeleteWithClient(svc)

	if err := batcher.Delete(aws.BackgroundContext(), iter); err != nil {
		return err
	}

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(s.uri.Path + "/"),
	})
	if err != nil {
		return err
	}

	// Create the config db
	tmpDir, err := ioutil.TempDir(s.workPath, "init")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	config := NewConfig(filepath.Join(tmpDir, "config.db"))

	err = config.Set("name", s.name)
	if err != nil {
		return err
	}

	// Upload the config db
	configDb, err := os.Open(filepath.Join(tmpDir, "config.db"))
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(s.session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
		Body:   configDb,
	})

	return err
}

func (s *S3Publisher) Update() error {
	tmpDir, err := ioutil.TempDir(s.workPath, "update")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	configDb, err := os.Create(filepath.Join(tmpDir, "config.db"))
	if err != nil {
		return err
	}
	defer configDb.Close()

	// Download the config db
	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(configDb, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", s.uri.Path))
	}

	// Modify config db
	config := NewConfig(filepath.Join(tmpDir, "config.db"))

	err = config.Set("name", s.name)
	if err != nil {
		return err
	}

	// Upload config db
	configDb.Seek(0, io.SeekStart)

	uploader := s3manager.NewUploader(s.session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
		Body:   configDb,
	})

	return err
}

func (s *S3Publisher) Channel(pkg string, channel string) error {
	for _, osarch := range zps.Platforms() {

		err := s.channel(osarch, pkg, channel)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *S3Publisher) Publish(pkgs ...string) error {
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

		pkgFiles, pkgs := FilterPackagesByArch(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := s.publish(osarch, pkgFiles, pkgs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *S3Publisher) channel(osarch *zps.OsArch, pkg string, channel string) error {
	tmpDir, err := ioutil.TempDir(s.workPath, "channel")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	// Download metadata db
	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(metadataDb, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", s.uri.Path))
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

		s.Emit("publisher.channel", pkg)
	}

	// Upload metadata
	metadataDb.Seek(0, io.SeekStart)

	uploader := s3manager.NewUploader(s.session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
		Body:   metadataDb,
	})

	return err
}

func (s *S3Publisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg) error {
	tmpDir, err := ioutil.TempDir(s.workPath, "publish")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	// Download metadata db
	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(metadataDb, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != "NoSuchKey" {
			return errors.New(fmt.Sprintf("unable to download: %s", s.uri.Path))
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

	rmFiles, err := repo.Prune(s.prune)
	if err != nil {
		return err
	}

	for _, r := range rmFiles {
		rejectIndex[r.FileName()] = true
	}

	svc := s3.New(s.session)

	if len(repo.Solvables()) > 0 {
		for _, file := range pkgFiles {
			if !rejectIndex[filepath.Base(file)] {
				s.Emit("publisher.publish", file)
				err = s.upload(file, path.Join(s.uri.Path, osarch.String(), filepath.Base(file)))
				if err != nil {
					return err
				}
			}
		}

		for _, pkg := range rmFiles {
			_, err = svc.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(s.uri.Host),
				Key:    aws.String(path.Join(s.uri.Path, osarch.String(), pkg.FileName())),
			})
			if err != nil {
				return err
			}
		}

		// TODO Rewrite instead of atomic update for now

		metadata.Empty()

		for _, pkg := range repo.Solvables() {
			metadata.Put(pkg.(*zps.Pkg))
		}

		// Upload metadata
		metadataUp, err := os.Open(metaPath)
		if err != nil {
			return err
		}
		defer metadataUp.Close()

		uploader := s3manager.NewUploader(s.session)

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
			Body:   metadataUp,
		})

	} else {
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String()) + "/"),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Publisher) upload(file string, dest string) error {
	src, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer src.Close()

	uploader := s3manager.NewUploader(s.session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(dest),
		Body:   src,
	})

	return err
}
