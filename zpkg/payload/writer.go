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
	"io/ioutil"
	"os"

	"github.com/dsnet/compress/bzip2"
)

type Writer struct {
	WorkPath string

	offset int64
	file   *os.File
}

func NewWriter(workPath string, offset int64) *Writer {
	writer := &Writer{}

	writer.WorkPath = workPath

	writer.offset = offset

	return writer
}

func (w *Writer) Put(path string) (int, int, string, error) {
	var err error
	if w.file == nil {
		w.file, err = ioutil.TempFile(w.WorkPath, "zpkgtmp")
		if err != nil {
			return 0, 0, "", err
		}
	}

	src, err := os.Open(path)
	if err != nil {
		return 0, 0, "", err
	}

	reader := bufio.NewReader(src)

	var writer *bzip2.Writer
	writer, _ = bzip2.NewWriter(w.file, &bzip2.WriterConfig{Level: 7})

	hasher := sha256.New()
	multi := io.MultiWriter(writer, hasher)

	_, err = io.Copy(multi, reader)
	if err != nil {
		return 0, 0, "", err
	}

	writer.Close()
	src.Close()

	currentOffset, err := w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, 0, "", err
	}

	offset := w.offset
	size := currentOffset - w.offset
	w.offset = currentOffset

	return int(offset), int(size), hex.EncodeToString(hasher.Sum(nil)), err
}

func (w *Writer) HasContents() bool {
	if w.file == nil {
		return false
	} else {
		return true
	}
}

func (w *Writer) Name() string {
	return w.file.Name()
}

func (w *Writer) Close() {
	w.file.Close()
}
