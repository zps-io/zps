package zpm

import (
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/solvent-io/zps/zps"
)

type LocalFetcher struct {
	uri   *url.URL
	cache *Cache
}

func NewLocalFetcher(uri *url.URL, cache *Cache) *FileFetcher {
	return &FileFetcher{uri, cache}
}

func (f *LocalFetcher) Refresh() error {
	return nil
}

func (f *LocalFetcher) Fetch(pkg *zps.Pkg) error {
	var err error
	packagefile := pkg.FileName()
	repofile := filepath.Join(f.uri.Path, packagefile)
	cachefile := f.cache.GetFile(packagefile)

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
