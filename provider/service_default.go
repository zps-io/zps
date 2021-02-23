/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package provider

import (
	"context"
	"fmt"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/action"
)

type ServiceDefault struct {
	*emission.Emitter
	service *action.Service

	phaseMap map[string]string
}

func NewServiceDefault(service action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &ServiceSystemD{emitter, service.(*action.Service), phaseMap}
}

func (s *ServiceDefault) Realize(ctx context.Context) error {
	switch s.phaseMap[Phase(ctx)] {
	default:
		s.Emit("action.info", fmt.Sprintf("%s %s", s.service.Type(), s.service.Key()))
		return nil
	}
}