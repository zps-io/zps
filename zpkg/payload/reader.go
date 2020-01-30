/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package payload

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/dsnet/compress/bzip2"
)

type Reader struct {
	WorkPath string
	Path     string

	offset int64
	file   *os.File
}

func NewReader(workPath string, path string, offset int64) *Reader {
	reader := &Reader{}

	reader.WorkPath = workPath
	reader.Path = path

	reader.offset = offset

	return reader
}

func (r *Reader) Get(path string, offset int64, size int64) (string, error) {
	var err error

	if r.file == nil {
		r.file, err = os.Open(r.Path)
		if err != nil {
			return "", err
		}
	}

	_, err = r.file.Seek(r.offset+offset, 0)
	if err != nil {
		return "", err
	}

	target, err := os.Create(path)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(r.file)

	var bzreader *bzip2.Reader

	bzreader, err = bzip2.NewReader(reader, &bzip2.ReaderConfig{})
	if err != nil {
		return "", err
	}

	writer := bufio.NewWriter(target)
	hasher := sha256.New()

	multi := io.MultiWriter(writer, hasher)

	_, err = io.CopyN(multi, bzreader, size)

	if err != nil {
		return "", err
	}

	writer.Flush()
	target.Close()

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) Close() {
	r.file.Close()
}
