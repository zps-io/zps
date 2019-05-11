/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zpkg

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/dsnet/compress/bzip2"
	"github.com/fezz-io/zps/action"
	"github.com/fezz-io/zps/zpkg/payload"
	"github.com/lunixbochs/struc"
)

type Reader struct {
	path     string
	workPath string

	Header   *Header
	Manifest *action.Manifest
	Payload  *payload.Reader
}

func NewReader(path string, workPath string) *Reader {
	reader := &Reader{}
	reader.path = path
	reader.workPath = workPath
	reader.Manifest = action.NewManifest()
	return reader
}

func (r *Reader) Read() error {
	file, err := os.Open(r.path)
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if r.workPath == "" {
		r.workPath = wd
	}

	reader := bufio.NewReader(file)

	r.Header = &Header{}
	struc.Unpack(reader, r.Header)

	// Check magic
	if string(r.Header.Magic[:6]) != Magic {
		return errors.New("does not appear to be a zpkg file")
	}

	var cManifestBytes bytes.Buffer
	_, err = io.CopyN(&cManifestBytes, reader, int64(r.Header.ManifestLength))
	if err != nil {
		return err
	}

	bzreader, err := bzip2.NewReader(&cManifestBytes, &bzip2.ReaderConfig{})
	if err != nil {
		return err
	}

	var manifestBytes bytes.Buffer
	writer := io.Writer(&manifestBytes)

	_, err = io.Copy(writer, bzreader)
	if err != nil {
		return err
	}

	err = r.Manifest.Load(manifestBytes.String())
	if err != nil {
		return err
	}

	file.Close()

	// TODO get byte size of header instead of just setting it
	offset := int64(r.Header.ManifestLength + 12)
	r.Payload = payload.NewReader(r.workPath, r.path, offset)
	return err
}

func (r *Reader) Close() {
	r.Payload.Close()
}
