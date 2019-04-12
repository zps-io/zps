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
	"io"
	"os"

	"github.com/dsnet/compress/bzip2"
	"github.com/lunixbochs/struc"
	"github.com/fezz-io/zps/action"
	"github.com/fezz-io/zps/zpkg/payload"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(filename string, header *Header, manifest *action.Manifest, payload *payload.Writer) error {
	var manifestBuffer bytes.Buffer

	// compress manifest
	bzw, _ := bzip2.NewWriter(&manifestBuffer, &bzip2.WriterConfig{Level: 7})

	if _, err := io.WriteString(bzw, manifest.ToJson()); err != nil {
		return err
	}

	if err := bzw.Close(); err != nil {
		return err
	}

	// Finalize Header
	header.ManifestLength = uint32(manifestBuffer.Len())

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	// Writer header
	struc.Pack(writer, header)
	writer.Flush()

	// Write manifest
	writer.Write(manifestBuffer.Bytes())
	writer.Flush()

	// Finish Payload
	if payload.HasContents() {
		payloadName := payload.Name()
		payload.Close()

		// Copy Payload to zpkg file
		payloadTmpFile, err := os.Open(payloadName)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(payloadTmpFile)
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
		writer.Flush()

		payloadTmpFile.Close()
		file.Close()

		// Cleanup
		err = os.Remove(payloadName)
		if err != nil {
			return err
		}
	}

	return err
}
