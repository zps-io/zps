/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zpm

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fezz-io/zps/phase"

	"time"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/action"
	"github.com/fezz-io/zps/provider"
	"github.com/fezz-io/zps/zpkg"
	"github.com/fezz-io/zps/zps"
	"github.com/segmentio/ksuid"
	"golang.org/x/net/context"
)

type Transaction struct {
	*emission.Emitter

	targetPath string
	cache      *Cache
	state      *State

	solution *zps.Solution
	readers  map[string]*zpkg.Reader

	id   ksuid.KSUID
	date time.Time
}

func NewTransaction(emitter *emission.Emitter, targetPath string, cache *Cache, state *State) *Transaction {
	return &Transaction{emitter, targetPath, cache, state, nil, nil, ksuid.New(), time.Now()}
}

func (t *Transaction) Realize(solution *zps.Solution) error {
	t.solution = solution
	t.readers = make(map[string]*zpkg.Reader)

	err := t.loadReaders()
	if err != nil {
		return err
	}

	err = t.solutionConflicts()
	if err != nil {
		return err
	}

	err = t.imageConflicts()
	if err != nil {
		return err
	}

	operations, err := t.solution.Graph()
	if err != nil {
		return err
	}

	for _, operation := range operations {
		switch operation.Operation {
		case "remove":
			t.Emit("transaction.remove", fmt.Sprint("removing ", operation.Package.Id()))
			err = t.remove(operation.Package)
			if err != nil {
				return err
			}
		case "install":
			// check if another version is installed and remove
			lookup, err := t.state.Packages.Get(operation.Package.Name())
			if err != nil {
				return err
			}

			if lookup != nil {
				lns, err := zps.NewPkgFromManifest(lookup)
				if err != nil {
					return err
				}

				t.Emit("transaction.remove", fmt.Sprint("removing ", lns.Id()))
				err = t.remove(operation.Package)
				if err != nil {
					return err
				}

				err = t.state.Transactions.Put(t.id.String(), lns.Id(), "remove", &t.date)
				if err != nil {
					return err
				}
			}

			t.Emit("transaction.install", fmt.Sprint("installing ", operation.Package.Id()))
			err = t.install(operation.Package)
			if err != nil {
				return err
			}
		}

		if operation.Operation != "noop" {
			err = t.state.Transactions.Put(t.id.String(), operation.Package.Id(), operation.Operation, &t.date)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Transaction) loadReaders() error {
	var err error

	// Read Manifests
	for _, operation := range t.solution.Operations() {
		if operation.Operation == "install" {
			reader := zpkg.NewReader(t.cache.GetFile(operation.Package.FileName()), "")

			err = reader.Read()
			if err != nil {
				return err
			}

			pkg, err := zps.NewPkgFromManifest(reader.Manifest)
			if err != nil {
				return err
			}

			t.readers[pkg.Name()] = reader
		}
	}

	return err
}

func (t *Transaction) solutionConflicts() error {
	var err error
	var fsActions action.Actions
	lookup := make(map[action.Action]*zps.Pkg)

	for _, reader := range t.readers {
		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		actions := reader.Manifest.Section("Dir", "File", "SymLink")

		// build lookup index, TODO revisit this
		for _, act := range actions {
			lookup[act] = pkg
		}

		fsActions = append(fsActions, actions...)
	}

	sort.Sort(fsActions)
	for index, act := range fsActions {
		prev := index - 1
		if prev != -1 {
			if act.Key() == fsActions[prev].Key() && act.Type() != "Dir" && fsActions[prev].Type() != "Dir" {
				return errors.New(fmt.Sprint(
					"Package Conflicts:\n",
					lookup[fsActions[prev]].Name(), " ", strings.ToUpper(fsActions[prev].Type()), " => ", fsActions[prev].Key(), "\n",
					lookup[act].Name(), " ", strings.ToUpper(act.Type()), " => ", act.Key()))
			}
		}
	}

	return err
}

func (t *Transaction) imageConflicts() error {
	var err error

	for _, reader := range t.readers {

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		for _, action := range reader.Manifest.Section("Dir", "File", "SymLink") {
			fsEntries, err := t.state.Objects.Get(action.Key())

			if err != nil {
				return err
			}

			for _, entry := range fsEntries {
				if entry.Pkg != pkg.Name() && entry.Type != "Dir" && action.Type() != "Dir" {
					return errors.New(fmt.Sprint(
						entry.Type,
						" ",
						entry.Path,
						" from installed pkg ",
						entry.Pkg,
						" conflicts with candidate ",
						pkg.Name()))
				}
			}
		}
	}

	return err
}

func (t *Transaction) install(pkg zps.Solvable) error {
	reader := t.readers[pkg.Name()]

	// Setup context
	ctx := context.WithValue(context.Background(), "options", &provider.Options{TargetPath: t.targetPath})
	ctx = context.WithValue(ctx, "phase", phase.INSTALL)
	ctx = context.WithValue(ctx, "payload", reader.Payload)

	// Provider Factory
	factory := provider.DefaultFactory(t.Emitter)

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		return err
	}

	var contents action.Actions
	contents = reader.Manifest.Section("Dir", "File", "SymLink")

	sort.Sort(contents)

	for _, fsObject := range contents {
		err = factory.Get(fsObject).Realize(ctx)
		if err != nil {
			return err
		}
	}

	// Add this to the package db
	err = t.state.Packages.Put(pkg.Name(), reader.Manifest)
	if err != nil {
		return err
	}

	// Add all the fs object to the fs db
	for _, fsObject := range contents {
		err = t.state.Objects.Put(fsObject.Key(), pkg.Name(), fsObject.Type())
		if err != nil {
			return err
		}
	}

	// Add templates to the tpl db
	templates := reader.Manifest.Section("Template")

	for _, tpl := range templates {
		err = t.state.Templates.Put(pkg.Name(), tpl.(*action.Template))
		if err != nil {
			return err
		}
	}

	return err
}

func (t *Transaction) remove(pkg zps.Solvable) error {
	lookup, err := t.state.Packages.Get(pkg.Name())
	if err != nil {
		return err
	}

	if lookup != nil {
		// Setup context
		ctx := context.WithValue(context.Background(), "options", &provider.Options{TargetPath: t.targetPath})
		ctx = context.WithValue(ctx, "phase", phase.REMOVE)

		// Provider Factory
		factory := provider.DefaultFactory(t.Emitter)

		pkg, err := zps.NewPkgFromManifest(lookup)
		if err != nil {
			return err
		}

		var contents action.Actions
		contents = lookup.Section("Dir", "File", "SymLink")

		// Reverse the actionlist
		sort.Sort(sort.Reverse(contents))

		for _, fsObject := range contents {
			err = factory.Get(fsObject).Realize(ctx)
			if err != nil {
				return err
			}
		}

		// Remove from the package db
		err = t.state.Packages.Del(pkg.Name())
		if err != nil {
			return err
		}

		// Remove fs objects from fs db
		err = t.state.Objects.Del(pkg.Name())
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Remove an existing frozen entry
		err = t.state.Frozen.Del(pkg.Id())

		// Remove templates from tpl db
		err = t.state.Templates.Del(pkg.Name())
		if err != nil {
			return err
		}
	}

	return err
}
