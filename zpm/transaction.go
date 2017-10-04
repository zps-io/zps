package zpm

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/context"

	"github.com/chuckpreslar/emission"
	"github.com/solvent-io/zps/action"
	"github.com/solvent-io/zps/db"
	"github.com/solvent-io/zps/provider"
	"github.com/solvent-io/zps/zpkg"
	"github.com/solvent-io/zps/zps"
)

// Simplistic for now as this will need to handle the various URIs and
// engage the solver for dependency resolution in the future design
// will also require more advanced URI handling/lookup
// Also this is basically a big mess
type Transaction struct {
	*emission.Emitter

	TargetPath string
	Phase      string
	Uris       []*url.URL
	Packages   []string

	readers map[string]*zpkg.Reader
	lookup  map[action.Action]*zps.Pkg
}

func NewTransaction(targetPath string, phase string, uris ...*url.URL) *Transaction {
	return &Transaction{emission.NewEmitter(), targetPath, phase, uris, nil, make(map[string]*zpkg.Reader), make(map[action.Action]*zps.Pkg)}
}

func (t *Transaction) AddUri(uri *url.URL) *Transaction {
	t.Uris = append(t.Uris, uri)
	return t
}

func (t *Transaction) AddPackage(pkg string) *Transaction {
	t.Packages = append(t.Packages, pkg)
	return t
}

func (t *Transaction) Realize() error {
	err := t.load()
	if err != nil {
		return err
	}

	switch t.Phase {
	case "install":
		return t.install()
	case "remove":
		return t.remove()
	default:
		return nil
	}
}

func (t *Transaction) install() error {
	err := t.resolve()
	if err != nil {
		return err
	}

	err = t.fetch()
	if err != nil {
		return err
	}

	if len(t.readers) == 0 {
		t.Emit("warn", "After resolution no packages to install")
		return nil
	}

	t.Emit("info", fmt.Sprint("Installing ", len(t.readers), " package(s)"))

	for pkg, reader := range t.readers {
		err = t.removePkg(pkg, true)
		if err != nil {
			return err
		}

		err = t.installPkg(reader)
		if err != nil {
			return err
		}
	}

	return err
}

func (t *Transaction) remove() error {
	var err error
	t.Emit("info", fmt.Sprint("Removing ", len(t.Packages), " package(s)"))

	for _, pkg := range t.Packages {
		err = t.removePkg(pkg, false)
		if err != nil {
			return err
		}
	}

	return err
}

// Load readers eliminate dupes
func (t *Transaction) load() error {
	var err error

	// Read Manifests
	for _, uri := range t.Uris {
		reader := zpkg.NewReader(uri.Path, "")

		err = reader.Read()
		if err != nil {
			return err
		}

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		if val, ok := t.readers[pkg.Name()]; ok {
			prev, err := zps.NewPkgFromManifest(val.Manifest)
			if err != nil {
				return err
			}

			if pkg.Version().GT(prev.Version()) {
				t.readers[pkg.Name()] = reader
			}
		} else {
			t.readers[pkg.Name()] = reader
		}
	}

	return err
}

func (t *Transaction) resolve() error {
	var err error
	var fsActions action.Actions

	// Ensure there are not conflicting packages in the list
	for _, reader := range t.readers {
		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		actions := reader.Manifest.Section("dir", "file", "symlink")

		// build lookup index, TODO revisit this
		for _, action := range actions {
			t.lookup[action] = pkg
		}

		fsActions = append(fsActions, actions...)
	}

	sort.Sort(fsActions)
	for index, action := range fsActions {
		prev := index - 1
		if prev != -1 {
			if action.Key() == fsActions[prev].Key() && action.Type() != "dir" && fsActions[prev].Type() != "dir" {
				return errors.New(fmt.Sprint(
					"Package Conflicts:\n",
					t.lookup[fsActions[prev]].Name(), " ", strings.ToUpper(fsActions[prev].Type()), " => ", fsActions[prev].Key(), "\n",
					t.lookup[action].Name(), " ", strings.ToUpper(action.Type()), " => ", action.Key()))
			}
		}
	}

	// Check package database
	for _, reader := range t.readers {
		current, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		lookup, err := db.GetPackage(current.Name())
		if err != nil {
			return err
		}

		if lookup != nil {
			lns, err := zps.NewPkgFromManifest(lookup)
			if err != nil {
				return err
			}

			if lns.Version().GT(current.Version()) {
				t.Emit(
					"warn",
					fmt.Sprint(
						current.Name(),
						"@",
						lns.Version().String(),
						" > candidate ",
						current.Version().String(),
						" skipping ..."))

				reader.Close()
				delete(t.readers, current.Name())
			}

			if lns.Version().EXQ(current.Version()) {
				t.Emit(
					"warn",
					fmt.Sprint(
						current.Name(),
						"@",
						lns.Version().String(),
						" already installed skipping ..."))

				reader.Close()
				delete(t.readers, current.Name())
			}
		}
	}

	// Check file database
	for _, reader := range t.readers {

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		for _, action := range reader.Manifest.Section("dir", "file", "symlink") {
			fsEntry, err := db.GetFsEntry(action.Key())

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

// This will fetch uris if required
func (t *Transaction) fetch() error {

	return nil
}

func (t *Transaction) installPkg(reader *zpkg.Reader) error {
	ctx := action.GetContext(action.NewOptions(), reader.Manifest)
	ctx = context.WithValue(ctx, "payload", reader.Payload)
	ctx.Value("options").(*action.Options).TargetPath = t.TargetPath

	pkg, err := zps.NewPkgFromManifest(reader.Manifest)
	if err != nil {
		return err
	}

	t.Emit("info", fmt.Sprint("+ ", pkg.Name(), "@", pkg.Version().String()))

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
	err = db.PutPackage(pkg.Name(), reader.Manifest)
	if err != nil {
		return err
	}

	// Add all the fs object to the fs db
	for _, fsObject := range contents {
		err = db.PutFsEntry(fsObject.Key(), pkg.Name(), fsObject.Type())
		if err != nil {
			return err
		}
	}

	return err
}

func (t *Transaction) removePkg(pkg string, quiet bool) error {
	lookup, err := db.GetPackage(pkg)
	if err != nil {
		return err
	}

	if lookup != nil {
		ctx := action.GetContext(action.NewOptions(), lookup)
		ctx.Value("options").(*action.Options).TargetPath = t.TargetPath

		pkg, err := zps.NewPkgFromManifest(lookup)
		if err != nil {
			return err
		}

		t.Emit("info", fmt.Sprint("[red]- ", pkg.Name(), "@", pkg.Version().String()))

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
		err = db.DelPackage(pkg.Name())
		if err != nil {
			return err
		}

		// Remove fs objects from fs db
		for _, fsObject := range contents {
			err = db.DelFsEntry(fsObject.Key(), pkg.Name())
			if err != nil {
				return err
			}
		}
	} else {
		if quiet == false {
			t.Emit("warn", fmt.Sprint("~ ", pkg, " not installed"))
		}
	}

	return err
}
