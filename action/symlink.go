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

type SymLink struct {
	Path   string `json:"path" hcl:"path,label"`
	Owner  string `json:"owner" hcl:"owner,optional"`
	Group  string `json:"group" hcl:"group,optional"`
	Target string `json:"target" hcl:"target"`
}

func NewSymLink() *SymLink {
	return &SymLink{}
}

func (s *SymLink) Key() string {
	return s.Path
}

func (s *SymLink) Type() string {
	return "SymLink"
}

func (s *SymLink) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(s.Type()),
		"",
		s.Owner + ":" + s.Group,
		s.Path,
	}, "|")
}

func (s *SymLink) Id() string {
	return fmt.Sprint(s.Type(), ".", s.Key())
}

func (s *SymLink) Condition() *bool {
	return nil
}

func (s *SymLink) MayFail() bool {
	return false
}

func (s *SymLink) IsValid() bool {
	if s.Path != "" && s.Owner != "" && s.Group != "" && s.Target != "" {
		return true
	}

	return false
}
