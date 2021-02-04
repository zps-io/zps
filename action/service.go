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

type Service struct {
	Name   string `json:"name" hcl:"name,label"`
	Timer bool  `json:"timer" hcl:"timer,optional"`
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Key() string {
	return s.Name
}

func (s *Service) Type() string {
	return "Service"
}

func (s *Service) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(s.Type()),
		s.Name,
	}, "|")
}

func (s *Service) Id() string {
	return fmt.Sprint(s.Type(), ".", s.Key())
}

func (s *Service) Condition() *bool {
	return nil
}

func (s *Service) MayFail() bool {
	return false
}

func (s *Service) IsValid() bool {
	if s.Name != "" {
		return true
	}

	return false
}
