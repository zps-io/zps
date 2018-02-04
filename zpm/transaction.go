package zpm

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/provider"
	"github.com/solvent-io/zps/zpkg"
	"github.com/solvent-io/zps/zps"
	"golang.org/x/net/context"
)

type Transaction struct {
	*emission.Emitter

	targetPath string
	cache      *Cache
	db         *Db

	solution *zps.Solution
	readers  map[string]*zpkg.Reader
}

func NewTransaction(targetPath string, cache *Cache, db *Db) *Transaction {
	return &Transaction{emission.NewEmitter(), targetPath, cache, db, nil, nil}
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

	// TODO refactor into an ordered operation plan based on graph deps
	for _, operation := range t.solution.Operations() {
		switch operation.Operation {
		case "remove":
			t.Emit("remove", fmt.Sprint("[red]- removing ", operation.Package.Id()))
			err = t.remove(operation.Package)
			if err != nil {
				return err
			}
		case "install":
			// check if another version is installed and remove
			lookup, err := t.db.GetPackage(operation.Package.Name())
			if err != nil {
				return err
			}

			if lookup != nil {
				lns, err := zps.NewPkgFromManifest(lookup)
				if err != nil {
					return err
				}

				t.Emit("remove", fmt.Sprint("[red]- removing ", lns.Id()))
				err = t.remove(operation.Package)
				if err != nil {
					return err
				}
			}

			t.Emit("install", fmt.Sprint("+ installing ", operation.Package.Id()))
			err = t.install(operation.Package)
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

		actions := reader.Manifest.Section("dir", "file", "symlink")

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
			if act.Key() == fsActions[prev].Key() && act.Type() != "dir" && fsActions[prev].Type() != "dir" {
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

		for _, action := range reader.Manifest.Section("dir", "file", "symlink") {
			fsEntry, err := t.db.GetFsEntry(action.Key())

			if err != nil {
				return err
			}

			if fsEntry != nil && !fsEntry.Contains(pkg.Name()) && fsEntry.Type != "dir" && action.Type() != "dir" {
				return errors.New(fmt.Sprint(
					fsEntry.Type,
					" ",
					fsEntry.Path,
					" from installed pkg(s) ",
					fsEntry.ProvidedBy(),
					" conflicts with candidate ",
					pkg.Name()))
			}
		}
	}

	return err
}

func (t *Transaction) install(pkg zps.Solvable) error {
	reader := t.readers[pkg.Name()]
	ctx := action.GetContext(action.NewOptions(), reader.Manifest)
	ctx = context.WithValue(ctx, "payload", reader.Payload)
	ctx.Value("options").(*action.Options).TargetPath = t.targetPath

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		return err
	}

	var contents action.Actions
	contents = reader.Manifest.Section("dir", "file", "symlink")

	sort.Sort(contents)

	for _, fsObject := range contents {
		err = provider.Get(fsObject).Realize("install", ctx)
		if err != nil {
			return err
		}
	}

	// Add this to the package db
	err = t.db.PutPackage(pkg.Name(), reader.Manifest)
	if err != nil {
		return err
	}

	// Add all the fs object to the fs db
	for _, fsObject := range contents {
		err = t.db.PutFsEntry(fsObject.Key(), pkg.Name(), fsObject.Type())
		if err != nil {
			return err
		}
	}

	return err
}

func (t *Transaction) remove(pkg zps.Solvable) error {
	lookup, err := t.db.GetPackage(pkg.Name())
	if err != nil {
		return err
	}

	if lookup != nil {
		ctx := action.GetContext(action.NewOptions(), lookup)
		ctx.Value("options").(*action.Options).TargetPath = t.targetPath

		pkg, err := zps.NewPkgFromManifest(lookup)
		if err != nil {
			return err
		}

		var contents action.Actions
		contents = lookup.Section("dir", "file", "symlink")

		// Reverse the actionlist
		sort.Sort(sort.Reverse(contents))

		for _, fsObject := range contents {
			err = provider.Get(fsObject).Realize("remove", ctx)
			if err != nil {
				return err
			}
		}

		// Remove from the package db
		err = t.db.DelPackage(pkg.Name())
		if err != nil {
			return err
		}

		// Remove fs objects from fs db
		for _, fsObject := range contents {
			err = t.db.DelFsEntry(fsObject.Key(), pkg.Name())
			if err != nil {
				return err
			}
		}
	}

	return err
}
