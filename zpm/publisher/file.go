package publisher

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"io"
	"io/ioutil"

	"github.com/solvent-io/zps/zpkg"
	"github.com/solvent-io/zps/zps"
	"fmt"
)

type FilePublisher struct {
	uri   *url.URL
	prune int
}

func NewFilePublisher(uri *url.URL, prune int) *FilePublisher {
	return &FilePublisher{uri, prune}
}

func (f *FilePublisher) Init() error {
	os.MkdirAll(f.uri.Path, os.FileMode(0750))

	for _, osarch := range zps.Platforms() {
		os.RemoveAll(filepath.Join(f.uri.Path, osarch.String()))
	}

	return nil
}

func (f *FilePublisher) Publish(pkgs ...string) error {
	zpkgs := make(map[string]*zps.Pkg)
	for _, file := range pkgs {
		reader := zpkg.NewReader(file, "")

		err := reader.Read()
		if err != nil {
			return err
		}

		pkg, err := zps.NewPkgFromManifest(reader.Manifest)
		if err != nil {
			return err
		}

		zpkgs[file] = pkg
	}

	for _, osarch := range zps.Platforms() {

		pkgFiles, pkgs := f.filter(osarch, zpkgs)
		if len(pkgFiles) > 0 {
			err := f.publish(osarch, pkgFiles, pkgs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *FilePublisher) publish(osarch *zps.OsArch, pkgFiles []string, zpkgs []*zps.Pkg) error {
	var err error

	lockfile := filepath.Join(f.uri.Path, osarch.String(), ".lock")
	packagesfile := filepath.Join(f.uri.Path, osarch.String(), "packages.json")
	meta := &RepoMeta{}

	if _, err = os.Stat(lockfile); !os.IsNotExist(err) {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	} else {
		os.OpenFile(lockfile, os.O_RDONLY|os.O_CREATE, 0640)
		defer os.Remove(lockfile)
	}

	pkgsbytes, err := ioutil.ReadFile(packagesfile)

	if err == nil {
		err = meta.Load(pkgsbytes)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	meta.Add(zpkgs...)

	rmFiles, err := meta.Prune(f.prune)
	if err != nil {
		return err
	}

	if len(meta.Repo.Solvables) > 0 {
		json, err := meta.Json()
		if err != nil {
			return err
		}

		os.Mkdir(filepath.Join(f.uri.Path, osarch.String()), 0750)

		for _, file := range pkgFiles {
			fmt.Println(file)
			err = f.upload(file, filepath.Join(f.uri.Path, osarch.String(), filepath.Base(file)))
			if err != nil {
				return err
			}
		}

		for _, file := range rmFiles {
			os.Remove(filepath.Join(f.uri.Path, osarch.String(), file))
		}

		err = ioutil.WriteFile(packagesfile, json, 0640)
		if err != nil {
			return err
		}
	} else {
		os.Remove(packagesfile)
	}

	return nil
}

func (f *FilePublisher) filter(osarch *zps.OsArch, zpkgs map[string]*zps.Pkg) ([]string, []*zps.Pkg) {
	var files []string
	var pkgs []*zps.Pkg

	for k, v := range zpkgs {
		if v.Os() == osarch.Os && v.Arch() == osarch.Arch {
			files = append(files, k)
			pkgs = append(pkgs, zpkgs[k])
		}
	}

	return files, pkgs
}

func (f *FilePublisher) upload(file string, dest string) error {
	s, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}

	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}

	return d.Close()
}
