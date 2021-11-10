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
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/chuckpreslar/emission"

	"github.com/fezz-io/zps/sec"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
)

type S3Publisher struct {
	*emission.Emitter

	security Security

	workPath string

	uri   *url.URL
	name  string
	prune int

	session *session.Session

	lockUri *url.URL
}

func NewS3Publisher(emitter *emission.Emitter, security Security, workPath string, uri *url.URL, name string, prune int, lockUri *url.URL) *S3Publisher {
	sess := session.Must(session.NewSession())

	user := uri.User.Username()
	password, _ := uri.User.Password()

	if user != "" && password != "" {
		sess.Config.Credentials = credentials.NewStaticCredentials(user, password, "")
	}

	region, err := s3manager.GetBucketRegion(context.Background(), sess, uri.Host, "us-west-2")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			emitter.Emit("error", fmt.Sprintf("unable to find bucket %s's region not found\n", uri.Host))
		}

		return nil
	}

	sess.Config.Region = aws.String(region)

	return &S3Publisher{emitter, security, workPath, uri, name, prune, sess, lockUri}
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

	configPath := filepath.Join(tmpDir, "config.db")
	sigPath := filepath.Join(tmpDir, "config.sig")
	config := NewConfig(configPath)

	err = config.Set("name", s.name)
	if err != nil {
		return err
	}

	// Upload the config db
	configDb, err := os.Open(configPath)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.Concurrency = 3
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
		Body:   configDb,
	})

	if err != nil {
		return fmt.Errorf("unable to upload config.db for repo: %s, err: %s", s.uri.Path, err.Error())
	}

	// Sign and upload
	keyPair, err := s.security.KeyPair(PublisherFromUri(s.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		s.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(s.uri)))
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

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, "config.sig")),
			Body:   configSig,
		})

		if err != nil {
			return fmt.Errorf("unable to upload config.sig for repo: %s, err: %s", s.uri.Path, err.Error())
		}
	}

	return err
}

func (s *S3Publisher) Update() error {
	tmpDir, err := ioutil.TempDir(s.workPath, "update")
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

	// Download the config db
	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(configDb, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
	})
	if err != nil {
		return fmt.Errorf("unable to download: %s, err: %s", s.uri.Path, err.Error())
	}

	// Modify config db
	config := NewConfig(configPath)

	err = config.Set("name", s.name)
	if err != nil {
		return err
	}

	// Upload config db
	configDb.Seek(0, io.SeekStart)

	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.Concurrency = 3
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
		Body:   configDb,
	})

	if err != nil {
		return fmt.Errorf("unable to upload config.db for repo: %s, err: %s", s.uri.Path, err.Error())
	}

	// Sign and upload
	keyPair, err := s.security.KeyPair(PublisherFromUri(s.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		s.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(s.uri)))
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

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, "config.sig")),
			Body:   configSig,
		})

		if err != nil {
			return fmt.Errorf("unable to upload config.sig for repo: %s, err: %s", s.uri.Path, err.Error())
		}
	}

	return err
}

func (s *S3Publisher) Channel(pkg string, channel string) error {
	keyPair, err := s.security.KeyPair(PublisherFromUri(s.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		s.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(s.uri)))
	}

	for _, osarch := range zps.Platforms() {
		err := s.channel(osarch, pkg, channel, keyPair)
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

	keyPair, err := s.security.KeyPair(PublisherFromUri(s.uri))
	if err != nil {
		return err
	}

	if keyPair == nil {
		s.Emitter.Emit("publisher.warn", fmt.Sprintf("No keypair found for publisher %s, not signing.", PublisherFromUri(s.uri)))
	}

	for _, osarch := range zps.Platforms() {
		pkgFiles, pkgs := FilterPackagesByArch(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := s.publish(osarch, pkgFiles, pkgs, keyPair)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *S3Publisher) channel(osarch *zps.OsArch, pkg string, channel string, keyPair *KeyPairEntry) error {
	retries := 10

	tmpDir, err := ioutil.TempDir(s.workPath, "channel")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	locker := NewLocker(s.lockUri)

	eTag, err := locker.LockWithEtag()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", s.name, err.Error())
	}

	defer locker.UnlockWithEtag(&eTag)

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	// Download metadata db
	client := s3manager.NewDownloader(s.session)

	for {
		size, err := client.Download(metadataDb, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
		})
		if err != nil {
			return fmt.Errorf("unable to download metadata file: %s, err: %s", s.uri.Path, err.Error())
		}

		metadataDbData := make([]byte, size)

		_, err = metadataDb.Read(metadataDbData)
		if err != nil {
			return fmt.Errorf("unable to read downloaded file: %s", s.uri.Path)
		}

		actualETag := md5.Sum(metadataDbData)

		// if eTag is empty, it means locker method doesn't support
		// storing attribute or doesn't contain previous eTag
		// we will break from cycle right after
		if len(eTag) > 0 && actualETag != eTag {
			retries -= 1
			if retries == 0 {
				return fmt.Errorf("object %q has eTag mismatch: want %q, got %q", s.uri.Path, eTag, actualETag)
			}
			time.Sleep(6 * time.Second)
			continue
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

		s.Emit("publisher.channel", pkg)
	}

	// Upload metadata
	metadataDb.Seek(0, io.SeekStart)

	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.Concurrency = 3
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
		Body:   metadataDb,
	})

	if err != nil {
		return fmt.Errorf("unable to upload new metadata file: %s", s.uri.Path)
	}

	medatabaDbData := make([]byte, 0)
	_, err = metadataDb.Read(medatabaDbData)
	if err != nil {
		return fmt.Errorf("unable to read downloaded file: %s", s.uri.Path)
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

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.sig")),
			Body:   metadataSig,
		})

		if err != nil {
			return err
		}
	}

	return err
}

func (s *S3Publisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg, keyPair *KeyPairEntry) error {
	retries := 10

	tmpDir, err := ioutil.TempDir(s.workPath, "publish")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	metaPath := filepath.Join(tmpDir, "metadata.db")
	sigPath := filepath.Join(tmpDir, "metadata.sig")

	locker := NewLocker(s.lockUri)

	eTag, err := locker.LockWithEtag()
	if err != nil {
		return fmt.Errorf("repository: %s is locked by another process, error: %s", s.name, err.Error())
	}

	defer locker.UnlockWithEtag(&eTag)

	metadataDb, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer metadataDb.Close()

	// Download metadata db
	client := s3manager.NewDownloader(s.session)

	for {
		_, err := client.Download(metadataDb, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
		})
		if err != nil {
			return fmt.Errorf("unable to download metadata file: %s, err: %s", s.uri.Path, err.Error())
		}

		medatabaDbData := make([]byte, 0)

		_, err = metadataDb.Read(medatabaDbData)
		if err != nil {
			return fmt.Errorf("unable to read downloaded metadata file: %s, err: %s", s.uri.Path, err.Error())
		}

		actualETag := md5.Sum(medatabaDbData)

		// if eTag is empty, it means locker method doesn't support
		// storing attribute or doesn't contain previous eTag
		// we will break from cycle right after
		if len(eTag) > 0 && actualETag != eTag {
			retries -= 1
			if retries == 0 {
				return fmt.Errorf("object %q has eTag mismatch: want %q, got %q", s.uri.Path, string(eTag[:]), string(actualETag[:]))
			}
			time.Sleep(6 * time.Second)
			continue
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
				s.Emit("spin.start", fmt.Sprintf("publishing: %s", file))

				err = s.upload(file, path.Join(s.uri.Path, osarch.String(), filepath.Base(file)))
				if err != nil {
					s.Emit("spin.error", fmt.Sprintf("failed: %s", file))
					return err
				}

				s.Emit("spin.success", fmt.Sprintf("published: %s", file))
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

		uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
			u.Concurrency = 3
		})

		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.db")),
			Body:   metadataUp,
		})

		if err != nil {
			return fmt.Errorf("unable to upload new metadata file: %s", s.uri.Path)
		}

		metadataDbData := make([]byte, 0)
		_, err = metadataDb.Read(metadataDbData)
		if err != nil {
			return fmt.Errorf("unable to read new metadata file: %s, err: %s", s.uri.Path, err.Error())
		}

		// Updated eTag will go to the same defer function
		eTag = md5.Sum(metadataDbData)

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

			configSig, err := os.Open(sigPath)
			if err != nil {
				return err
			}

			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(s.uri.Host),
				Key:    aws.String(path.Join(s.uri.Path, osarch.String(), "metadata.sig")),
				Body:   configSig,
			})

			if err != nil {
				return err
			}
		}
	} else {
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, osarch.String()) + "/"),
		})
		if err != nil {
			return err
		}
	}

	return err
}

func (s *S3Publisher) upload(file string, dest string) error {
	src, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer src.Close()

	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.Concurrency = 3
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(dest),
		Body:   src,
	})

	return err
}
