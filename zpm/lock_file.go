package zpm

import (
	"log"
	"net/url"

	"github.com/nightlyone/lockfile"
)

type FileLocker struct {
	lockfile *lockfile.Lockfile
}

func NewFileLocker(uri *url.URL) *FileLocker {
	lock, err := lockfile.New(uri.Path)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return &FileLocker{
		lockfile: &lock,
	}
}

func (f *FileLocker) Lock() error {
	return f.lockfile.TryLock()
}

func (f *FileLocker) Unlock() error {
	return f.lockfile.Unlock()
}
