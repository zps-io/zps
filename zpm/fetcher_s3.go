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
	d, err := os.Create(s.cache.GetConfig(s.uri.String()))
	if err != nil {
		return err
	}
	defer d.Close()

	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(path.Join(s.uri.Path, "config.db")),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", s.uri.Path))
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

	d, err := os.Create(s.cache.GetFile(pkg.FileName()))
	if err != nil {
		return err
	}
	defer d.Close()

	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(target),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", target))
	}

	return err
}

func (s *S3Fetcher) refresh(osarch *zps.OsArch) error {
	var err error
	target := path.Join(s.uri.Path, osarch.String(), "metadata.db")
	dest := s.cache.GetMeta(osarch.String(), s.uri.String())

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	client := s3manager.NewDownloader(s.session)

	_, err = client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(target),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() != "NoSuchKey" {
			return errors.New(fmt.Sprintf("unable to download: %s", target))
		} else {
			d.Close()
			os.Remove(dest)
		}
	}

	return nil
}
