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

type Signature struct {
	FingerPrint string `json:"fingerprint"`
	Algo        string `json:"algo"`
	Manifest    string `json:"manifest"`
}

func NewSignature() *File {
	return &File{}
}

func (s *Signature) Key() string {
	return s.FingerPrint
}

func (s *Signature) Type() string {
	return "Signature"
}

func (s *Signature) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(s.Type()),
		s.FingerPrint,
	}, "|")
}

func (s *Signature) Id() string {
	return fmt.Sprint(s.Type(), ".", s.Key())
}

func (s *Signature) Condition() *bool {
	return nil
}

func (s *Signature) MayFail() bool {
	return false
}

func (s *Signature) IsValid() bool {
	if s.FingerPrint != "" && s.Algo != "" && s.Manifest != "" {
		return true
	}

	return false
}
