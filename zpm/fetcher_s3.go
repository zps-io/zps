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
	uri    *url.URL
	cache  *Cache
	client *s3manager.Downloader
}

func NewS3Fetcher(uri *url.URL, cache *Cache) *S3Fetcher {
	sess := session.Must(session.NewSession())

	user := uri.User.Username()
	password, _ := uri.User.Password()

	sess.Config.Credentials = credentials.NewStaticCredentials(user, password, "")

	region, err := s3manager.GetBucketRegion(context.Background(), sess, uri.Host, "us-west-2")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			fmt.Fprintf(os.Stderr, "unable to find bucket %s's region not found\n", uri.Host)
		}
		return nil
	}

	sess.Config.Credentials = credentials.NewStaticCredentials(user, password, "")
	sess.Config.Region = aws.String(region)

	client := s3manager.NewDownloader(sess)

	return &S3Fetcher{uri, cache, client}
}

func (s *S3Fetcher) Refresh() error {
	configUri, _ := url.Parse(s.uri.String())
	configUri.Path = path.Join(configUri.Path, "config.db")

	d, err := os.Create(s.cache.GetConfig(s.uri.String()))
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = s.client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(s.uri.Path),
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

	fileUri, _ := url.Parse(s.uri.String())
	fileUri.Path = path.Join(fileUri.Path, osarch.String(), pkg.FileName())

	d, err := os.Create(s.cache.GetFile(pkg.FileName()))
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = s.client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(fileUri.Path),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", fileUri.Path))
	}

	return nil
}

func (s *S3Fetcher) refresh(osarch *zps.OsArch) error {
	var err error

	metadataUri, _ := url.Parse(s.uri.String())
	metadataUri.Path = path.Join(metadataUri.Path, osarch.String(), "metadata.db")

	d, err := os.Create(s.cache.GetMeta(osarch.String(), s.uri.String()))
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = s.client.Download(d, &s3.GetObjectInput{
		Bucket: aws.String(s.uri.Host),
		Key:    aws.String(metadataUri.Path),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("unable to download: %s", metadataUri.Path))
	}

	return nil
}
