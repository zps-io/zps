package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/solvent-io/zps/zps"
)

type FileFetcher struct {
	uri       *url.URL
	cachePath string
}

func NewFileFetcher(uri *url.URL, cachePath string) *FileFetcher {
	return &FileFetcher{uri, cachePath}
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
	lockfile := filepath.Join(f.uri.Path, osarch.String(), ".lock")
	packagefile := filepath.Join(f.uri.Path, osarch.String(), zps.ZpkgFileName(pkg.Name(), pkg.Version().String(), pkg.Os(), pkg.Arch()))
	cachefile := filepath.Join(f.cachePath, zps.ZpkgFileName(pkg.Name(), pkg.Version().String(), pkg.Os(), pkg.Arch()))

	if _, err = os.Stat(lockfile); !os.IsNotExist(err) {
		return errors.New("Repository: " + f.uri.String() + " " + osarch.String() + " is locked by another process")
	} else {
		os.OpenFile(lockfile, os.O_RDONLY|os.O_CREATE, 0640)
		defer os.Remove(lockfile)
	}

	s, err := os.OpenFile(packagefile, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}
	defer s.Close()

	if _, err := os.Stat(cachefile); os.IsNotExist(err) {

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

	lockfile := filepath.Join(f.uri.Path, osarch.String(), ".lock")
	packagesfile := filepath.Join(f.uri.Path, osarch.String(), "packages.json")
	meta := &zps.RepoMeta{}

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

		// TODO migrate this functionality
		hasher := sha256.New()
		hasher.Write([]byte(f.uri.String()))

		repoId := hex.EncodeToString(hasher.Sum(nil))

		s, err := os.OpenFile(packagesfile, os.O_RDWR|os.O_CREATE, 0640)
		if err != nil {
			return err
		}
		defer s.Close()

		d, err := os.Create(filepath.Join(f.cachePath, fmt.Sprint(repoId, ".", osarch.String(), ".packages.json")))
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
