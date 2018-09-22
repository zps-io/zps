/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zpkg

/*
	Versions: 0
	Compression: 0 bzip2
*/

const (
	Magic       string = "zpkg66"
	Version     uint8  = 0
	Compression uint8  = 0

	DefaultZpfPath   = "Zpkgfile"
	DefaultTargetDir = "proto"
)

type Header struct {
	Magic       [6]byte `struc:"little"`
	Version     uint8
	Compression uint8

	ManifestLength uint32
}

func NewHeader(version uint8, compression uint8) *Header {
	header := &Header{Version: version, Compression: compression}
	copy(header.Magic[:], Magic)
	return header
}
