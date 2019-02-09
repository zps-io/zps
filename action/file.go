/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package action

import (
	"fmt"
	"strings"
)

type File struct {
	Path  string `json:"path" hcl:"path,label"`
	Owner string `json:"owner" hcl:"owner,optional"`
	Group string `json:"group" hcl:"group,optional"`
	Mode  string `json:"mode" hcl:"mode,optional"`

	Digest string `json:"digest"`
	Offset int    `json:"offset"`
	Csize  int    `json:"csize"`
	Size   int    `json:"size"`
}

func NewFile() *File {
	return &File{}
}

func (f *File) Key() string {
	return f.Path
}

func (f *File) Type() string {
	return "File"
}

func (f *File) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(f.Type()),
		f.Mode,
		f.Owner + ":" + f.Group,
		f.Path,
	}, "|")
}

func (f *File) Id() string {
	return fmt.Sprint(f.Type(), ".", f.Key())
}

func (f *File) Condition() *bool {
	return nil
}

func (f *File) MayFail() bool {
	return false
}

func (f *File) IsValid() bool {
	if f.Path != "" && f.Owner != "" && f.Group != "" && f.Mode != "" {
		return true
	}

	return false
}
