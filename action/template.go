/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2020 Zachary Schneider
 */

package action

import (
	"fmt"
	"strings"
)

type Template struct {
	Name   string `json:"name" hcl:"path,label"`
	Source string `json:"path" hcl:"source"`
	Output string `json:"path" hcl:"output"`

	Owner string `json:"owner" hcl:"owner,optional"`
	Group string `json:"group" hcl:"group,optional"`
	Mode  string `json:"mode" hcl:"mode,optional"`
}

func NewTemplate() *Template {
	return &Template{}
}

func (t *Template) Key() string {
	return t.Name
}

func (t *Template) Type() string {
	return "Template"
}

func (t *Template) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(t.Type()),
		t.Name,
		t.Source,
		t.Output,
	}, "|")
}

func (t *Template) Id() string {
	return fmt.Sprint(t.Type(), ".", t.Key())
}

func (t *Template) Condition() *bool {
	return nil
}

func (t *Template) MayFail() bool {
	return false
}

func (t *Template) IsValid() bool {
	if t.Name != "" && t.Source != "" && t.Output != "" && t.Owner != "" && t.Group != "" && t.Mode != "" {
		return true
	}

	return false
}
