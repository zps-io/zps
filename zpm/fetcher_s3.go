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
	"os"
	"strings"

	"github.com/chuckpreslar/emission"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"net/url"
	"path"

	"github.com/fezz-io/zps/zps"
)

type S3Fetcher struct {
	uri *url.URL

	cache    *Cache
	security Security

	session *session.Session
}

func NewS3Fetcher(uri *url.URL, cache *Cache, security Security) *S3Fetcher {
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
			fmt.Fprintf(os.Stderr, "unable to find bucket %s's region not found\n", uri.Host)
		}
		return nil
	}

	sess.Config.Region = aws.String(region)

	return &S3Fetcher{uri, cache, security, sess}
}

func (s *S3Fetcher) Refresh() error {
	dst, err := os.OpenFile(s.cache.GetConfig(s.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(dst, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
	})
	if err != nil {
		dst.Close()
		os.Remove(s.cache.GetConfig(s.uri.String()))

		return errors.New(fmt.Sprintf("refresh failed: %s", s.uri.String()))
	}

	if s.security.Mode() != SecurityModeNone {
		sdst, err := os.OpenFile(s.cache.GetConfigSig(s.uri.String()), os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer sdst.Close()

		_, err = client.Download(sdst, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(path.Join(s.uri.Path, "config.sig")),
		})
		if err != nil {
			sdst.Close()
			os.Remove(s.cache.GetConfigSig(s.uri.String()))

			return errors.New(fmt.Sprintf("refresh failed: %s", s.uri.String()))
		}

		// Validate config signature
		err = ValidateFileSignature(s.security, s.cache.GetConfig(s.uri.String()), s.cache.GetConfigSig(s.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(s.cache.GetConfig(s.uri.String()))
			os.Remove(s.cache.GetConfigSig(s.uri.String()))

			return err
		}
	}

	for _, osarch := range zps.Platforms() {
		err := s.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *S3Fetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}
	target := path.Join(s.uri.Path, osarch.String(), pkg.FileName())
	cacheFile := s.cache.GetFile(pkg.FileName())

	// Fetch package if not in cache
	if !s.cache.Exists(cacheFile) {
		dst, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer dst.Close()

		client := s3manager.NewDownloader(s.session)

		_, err = client.Download(dst, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(target),
		})
		if err != nil {
			return errors.New(fmt.Sprintf("unable to download: %s", target))
		}
	}

	// Validate pkg
	if s.security.Mode() != SecurityModeNone {
		err = ValidateZpkg(&emission.Emitter{}, s.security, cacheFile, true)
		if err != nil {
			os.Remove(cacheFile)

			return errors.New(fmt.Sprintf("failed to validate signature: %s", pkg.FileName()))
		}
	}

	return err
}

func (s *S3Fetcher) Keys() ([][]string, error) {
	client := s3.New(s.session)

	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.uri.Host),
		Prefix: aws.String(strings.TrimPrefix(s.uri.Path, "/") + "/"),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list items in bucket %q, %v", s.uri.Host, err)
	}

	var certs [][]string
	dl := s3manager.NewDownloader(s.session)

	for _, item := range resp.Contents {
		if !strings.Contains(*item.Key, ".pem") {
			continue
		}

		pem := &aws.WriteAtBuffer{}

		_, err = dl.Download(pem, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    item.Key,
		})

		cert := pem.Bytes()

		subject, publisher, err := s.security.Trust(&cert, "")
		if err != nil {
			return nil, err
		}

		certs = append(certs, []string{subject, publisher})
	}

	return certs, nil
}

func (s *S3Fetcher) refresh(osarch *zps.OsArch) error {
	var err error
	target := path.Join(s.uri.Path, osarch.String(), "metadata.db")
	dest := s.cache.GetMeta(osarch.String(), s.uri.String())

	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer dst.Close()

	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(dst, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(target),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != "NoSuchKey" {
			return errors.New(fmt.Sprintf("unable to download: %s", target))
		} else {
			dst.Close()
			os.Remove(dest)
			return nil
		}
	}

	if s.security.Mode() != SecurityModeNone {
		starget := path.Join(s.uri.Path, osarch.String(), "metadata.sig")
		sdest := s.cache.GetMetaSig(osarch.String(), s.uri.String())

		sdst, err := os.OpenFile(sdest, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer sdst.Close()

		_, err = client.Download(sdst, &s3.GetObjectInput{
			Bucket: aws.String(s.uri.Host),
			Key:    aws.String(starget),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() != "NoSuchKey" {
				return errors.New(fmt.Sprintf("unable to download: %s", target))
			} else {
				dst.Close()
				os.Remove(sdest)
			}
		}

		// Validate metadata signature
		err = ValidateFileSignature(s.security, s.cache.GetMeta(osarch.String(), s.uri.String()), s.cache.GetMetaSig(osarch.String(), s.uri.String()))
		if err != nil {
			// Remove the config and sig if validation fails
			os.Remove(s.cache.GetMeta(osarch.String(), s.uri.String()))
			os.Remove(s.cache.GetMetaSig(osarch.String(), s.uri.String()))

			return err
		}
	}

	return nil
}
