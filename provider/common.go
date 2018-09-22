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

	"github.com/solvent-io/zps/phase"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
)

type Provider interface {
	Realize(ctx context.Context) error
}

type Options struct {
	OutputPath string
	TargetPath string
	WorkPath   string
	CachePath  string

	Secure   bool
	Restrict bool

	Owner string
	Group string

	Debug   bool
	Verbose bool
}

type Factory struct {
	phaseMap    map[string]map[string]string
	providerMap map[string]func(action.Action, map[string]string, *emission.Emitter) Provider
	emitter     *emission.Emitter
}

func New(emitter *emission.Emitter) *Factory {
	return &Factory{
		make(map[string]map[string]string),
		make(map[string]func(action.Action, map[string]string, *emission.Emitter) Provider),
		emitter,
	}
}

// Need to add provider switching
// for now defaults will work on all OSs we care about
func (f *Factory) Get(ac action.Action) Provider {
	return f.providerMap[ac.Type()](ac, f.phaseMap[ac.Type()], f.emitter)
}

// Build phase map
func (f *Factory) On(action string, phase string, call string) *Factory {
	if f.phaseMap[action] == nil {
		f.phaseMap[action] = make(map[string]string)
	}
	f.phaseMap[action][phase] = call

	return f
}

// Register Provider
func (f *Factory) Register(provider string, newFunc func(action.Action, map[string]string, *emission.Emitter) Provider) *Factory {
	f.providerMap[provider] = newFunc

	return f
}

func Phase(ctx context.Context) string {
	return ctx.Value("phase").(string)
}

func Opts(ctx context.Context) *Options {
	return ctx.Value("options").(*Options)
}

func DefaultFactory(emitter *emission.Emitter) *Factory {
	factory := New(emitter)

	factory.
		Register("Dir", NewDirUnix).
		Register("File", NewFileUnix).
		Register("Requirement", NewRequirementDefault).
		Register("SymLink", NewSymLinkUnix).
		Register("Tag", NewTagDefault).
		Register("Zpkg", NewZpkgDefault)

	factory.
		On("Dir", phase.INSTALL, "install").
		On("Dir", phase.PACKAGE, "package").
		On("Dir", phase.REMOVE, "remove").
		On("File", phase.INSTALL, "install").
		On("File", phase.PACKAGE, "package").
		On("File", phase.REMOVE, "remove").
		On("Symlink", phase.INSTALL, "install").
		On("SymLink", phase.PACKAGE, "package").
		On("SymLink", phase.REMOVE, "remove")

	return factory
}
