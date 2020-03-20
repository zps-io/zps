/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package zps

import (
	"strings"

	"fmt"

	"github.com/fezz-io/zps/action"
)

type Pkg struct {
	reqs []*Requirement

	name      string
	version   *Version
	publisher string

	arch        string
	os          string
	summary     string
	description string

	channels []string

	location int
	priority int
}

type PkgEntry struct {
	Id string `storm:"id"`

	Requirements []*Requirement

	Name      string
	Version   *Version
	Publisher string

	Arch        string
	Os          string
	Summary     string
	Description string

	Channels []string
}

func NewPkg(name string, version string, publisher string, reqs []*Requirement, arch string, os string, summary string, description string) (*Pkg, error) {
	ver := &Version{}
	err := ver.Parse(version)
	if err != nil {
		return nil, err
	}
	return &Pkg{reqs, name, ver, publisher, arch, os, summary, description, nil, 0, 0}, nil
}

func NewPkgFromManifest(manifest *action.Manifest) (*Pkg, error) {
	pkg := &Pkg{}
	zpkg := manifest.Zpkg

	version := &Version{}
	err := version.Parse(zpkg.Version)
	if err != nil {
		return nil, err
	}

	pkg.name = zpkg.Name
	pkg.version = version
	pkg.publisher = zpkg.Publisher

	pkg.arch = zpkg.Arch
	pkg.os = zpkg.Os
	pkg.summary = zpkg.Summary
	pkg.description = zpkg.Description

	for _, raction := range manifest.Section("Requirement") {
		req := NewRequirement(raction.(*action.Requirement).Name, nil)

		if raction.(*action.Requirement).Version != "" {
			version := &Version{}
			err := version.Parse(raction.(*action.Requirement).Version)
			if err != nil {
				return nil, err
			}

			req.Version = version
		}

		switch raction.(*action.Requirement).Method {
		case "depends":
			req = req.Depends().Op(raction.(*action.Requirement).Operation)
		case "conflicts":
			req = req.Conflicts().Op(raction.(*action.Requirement).Operation)
		case "provides":
			req = req.Provides().ANY()
		}

		pkg.reqs = append(pkg.reqs, req)
	}

	return pkg, nil
}

func (p *Pkg) Id() string {
	return strings.Join([]string{p.name, p.version.String()}, "@")
}

func (p *Pkg) Name() string {
	return p.name
}

func (p *Pkg) Publisher() string {
	return p.publisher
}

func (p *Pkg) Summary() string {
	return p.summary
}

func (p *Pkg) Description() string {
	return p.description
}

func (p *Pkg) Version() *Version {
	return p.version
}

func (p *Pkg) Requirements() []*Requirement {
	return p.reqs
}

func (p *Pkg) Arch() string {
	return p.arch
}

func (p *Pkg) Os() string {
	return p.os
}

func (p *Pkg) Location() int {
	return p.location
}

func (p *Pkg) SetLocation(location int) {
	p.location = location
}

func (p *Pkg) Priority() int {
	return p.priority
}

func (p *Pkg) SetPriority(priority int) {
	p.priority = priority
}

func (p *Pkg) Satisfies(req *Requirement) bool {
	switch req.Operation {
	case 3:
		return true
	case 2:
		return p.version.EXQ(req.Version)
	case 1:
		return p.version.GTE(req.Version)
	case 0:
		return p.version.EQ(req.Version)
	case -1:
		return p.version.LTE(req.Version)
	}

	return false
}

func (p *Pkg) SetChannels(channels ...string) {
	if len(channels) == 0 {
		p.channels = nil
		return
	}

	if channels[0] == "" {
		p.channels = nil
		return
	}

	p.channels = append(p.channels, channels...)
}

func (p *Pkg) Channels() []string {
	return p.channels
}

func (p *Pkg) FileName() string {
	return fmt.Sprintf("%s@%s-%s-%s.zpkg", p.Name(), p.Version().String(), p.Os(), p.Arch())
}

func (p *Pkg) Columns() string {
	return strings.Join([]string{
		p.Name(),
		p.Id(),
		p.Summary(),
	}, "|")
}

func (p *Pkg) ToEntry() *PkgEntry {
	return &PkgEntry{
		Id:           p.Id(),
		Requirements: p.Requirements(),
		Name:         p.Name(),
		Version:      p.Version(),
		Publisher:    p.Publisher(),
		Arch:         p.Arch(),
		Os:           p.Os(),
		Summary:      p.Summary(),
		Description:  p.Description(),
		Channels:     p.Channels(),
	}
}

func (p *PkgEntry) ToPkg() *Pkg {
	return &Pkg{
		reqs:        p.Requirements,
		name:        p.Name,
		version:     p.Version,
		publisher:   p.Publisher,
		arch:        p.Arch,
		os:          p.Os,
		summary:     p.Summary,
		description: p.Description,
		channels:    p.Channels,
	}
}
