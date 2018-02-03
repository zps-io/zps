package zpm

import (
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nightlyone/lockfile"
	"github.com/solvent-io/zps/zps"
)

type FileFetcher struct {
	uri   *url.URL
	cache *Cache
}

func NewFileFetcher(uri *url.URL, cache *Cache) *FileFetcher {
	return &FileFetcher{uri, cache}
}

func (f *FileFetcher) Refresh() error {
	for _, osarch := range zps.Platforms() {
		err := f.refresh(osarch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *FileFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	osarch := &zps.OsArch{pkg.Os(), pkg.Arch()}
	packagefile := pkg.FileName()
	repofile := filepath.Join(f.uri.Path, osarch.String(), packagefile)
	cachefile := f.cache.GetFile(packagefile)

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	s, err := os.OpenFile(repofile, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	if !f.cache.Exists(cachefile) {
		d, err := os.Create(cachefile)
		if err != nil {
			return err
		}

		if _, err := io.Copy(d, s); err != nil {
			d.Close()
			return err
		}

		return d.Close()
	}

	return nil
}

func (f *FileFetcher) refresh(osarch *zps.OsArch) error {
	var err error

	packagesfile := filepath.Join(f.uri.Path, osarch.String(), "packages.json")
	meta := &zps.RepoMeta{}

	if _, err = os.Stat(filepath.Join(f.uri.Path, osarch.String())); os.IsNotExist(err) {
		return nil
	}

	lock, err := lockfile.New(filepath.Join(f.uri.Path, osarch.String(), ".lock"))
	if err != nil {
		return err
	}

	err = lock.TryLock()
	if err != nil {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	}
	defer lock.Unlock()

	pkgsbytes, err := ioutil.ReadFile(packagesfile)

	if err == nil {
		err = meta.Load(pkgsbytes)
		if err != nil {
			return err
		}

		s, err := os.OpenFile(packagesfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer s.Close()

		d, err := os.Create(f.cache.GetPackages(osarch.String(), f.uri.String()))
		if err != nil {
			return err
		}

		if _, err := io.Copy(d, s); err != nil {
			d.Close()
			return err
		}

		return d.Close()
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
