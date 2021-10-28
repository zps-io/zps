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
	"github.com/fezz-io/zps/systemd"
	"os"
	"path/filepath"

	"github.com/fezz-io/zps/action"
)

var UnitPath = "usr/lib/systemd/system"

type ServiceSystemD struct {
	*emission.Emitter
	service *action.Service

	phaseMap map[string]string
}

func NewServiceSystemD(service action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &ServiceSystemD{emitter, service.(*action.Service), phaseMap}
}

func (s *ServiceSystemD) Realize(ctx context.Context) error {
	switch s.phaseMap[Phase(ctx)] {
	case "configure":
		return s.configure(ctx)
	case "remove":
		return s.remove(ctx)
	default:
		s.Emit("action.info", fmt.Sprintf("%s %s", s.service.Type(), s.service.Key()))
		return nil
	}
}

func (s *ServiceSystemD) configure(ctx context.Context) error {
	options := Opts(ctx)

	source := filepath.Join(options.TargetPath, UnitPath, s.service.Name+".service")

	err := systemd.Link(s.service.Name+".service", source)
	if err != nil {
		return err
	}

	if s.service.Timer {
		source := filepath.Join(options.TargetPath, UnitPath, s.service.Name+".timer")

		err := systemd.Link(s.service.Name+".timer", source)
		if err != nil {
			return err
		}
	}

	err = systemd.Reload()
	if err != nil {
		return err
	}

	if s.service.Timer {
		err := systemd.EnableService(s.service.Name+".timer")
		if err != nil {
			return err
		}

		err = systemd.RestartService(s.service.Name+".timer")
		if err != nil {
			return err
		}
	} else {
		err := systemd.EnableService(s.service.Name+".service")
		if err != nil {
			return err
		}

		err = systemd.RestartService(s.service.Name+".service")
		if err != nil {
			return err
		}
	}

	s.Emit("action.info", fmt.Sprintf(
		"%s %s",
		s.service.Type(),
		s.service.Key(),
	))

	return nil
}

func (s *ServiceSystemD) remove(_ context.Context) error {
	err := systemd.UnLink(s.service.Name+".service")
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	if s.service.Timer {
		err := systemd.UnLink(s.service.Name+".timer")
		if err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
	}

	err = systemd.Reload()
	if err != nil {
		return err
	}

	return nil
}