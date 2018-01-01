package zpm

import (
	"github.com/boltdb/bolt"
	"errors"
	"strings"
	"bytes"
	"path/filepath"
	"time"
	"github.com/solvent-io/zps/action"
)

type Db struct {
	Path string
}

type FsEntry struct {
	Path string
	Type string
	Pkgs map[string]bool
}

func (f *FsEntry) Contains(pkg string) bool {
	return f.Pkgs[pkg]
}

func (f *FsEntry) ProvidedBy() string {
	pkgs := make([]string, 0, len(f.Pkgs))
	for k := range f.Pkgs {
		pkgs = append(pkgs, k)
	}

	return strings.Join(pkgs, ", ")
}

func (d *Db) getDb(name string) (*bolt.DB, error) {
	db, err := bolt.Open(filepath.Join(d.Path, name+".db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (d *Db) GetFsEntry(path string) (*FsEntry, error) {
	db, err := d.getDb("image")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("fs"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var fsEntry *FsEntry = &FsEntry{path, "", make(map[string]bool)}
	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("fs")).Cursor()

		prefix := []byte(path + "\x00")
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if fsEntry.Type == "" {
				fsEntry.Type = string(v)
			}
			if fsEntry.Type == string(v) {
				fsEntry.Pkgs[d.fsPkg(string(k))] = true
			} else {

				return errors.New("image db fs bucket inconsistent, must be repaired")
			}
		}

		return nil
	})

	if len(fsEntry.Pkgs) == 0 {
		fsEntry = nil
	}

	return fsEntry, err
}

func (d *Db) DelFsEntry(path string, pkg string) error {
	db, err := d.getDb("image")
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("fs"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fs"))
		v := b.Delete([]byte(d.fsKey(path, pkg)))

		if v != nil {
			return errors.New("FS object not recorded")
		}

		return nil
	})

	return err
}

func (d *Db) PutFsEntry(path string, pkg string, typ string) error {
	db, err := d.getDb("image")
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("fs"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("fs"))
		err := b.Put([]byte(d.fsKey(path, pkg)), []byte(typ))
		return err
	})

	return err
}

func (d *Db) fsKey(path string, pkg string) string {
	return strings.Join([]string{path, pkg}, "\x00")
}

func (d *Db) fsPkg(key string) string {
	return strings.Split(key, "\x00")[1]
}

func (d *Db) Packages() ([]*action.Manifest, error) {
	db, err := d.getDb("image")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var packages []*action.Manifest
	err = db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("packages"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var manifest *action.Manifest = action.NewManifest()
			err := manifest.Load(string(v))
			if err != nil {
				return err
			}

			packages = append(packages, manifest)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return packages, nil
}

func (d *Db) GetPackage(name string) (*action.Manifest, error) {
	db, err := d.getDb("image")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var pkg *action.Manifest
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("packages"))
		v := b.Get([]byte(name))

		if v != nil {
			pkg = action.NewManifest()
			err := pkg.Load(string(v))
			if err != nil {
				return err
			}
		}

		return nil
	})

	return pkg, err
}

func (d *Db) DelPackage(name string) error {
	db, err := d.getDb("image")
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("packages"))
		v := b.Delete([]byte(name))

		if v != nil {
			return errors.New("Package not installed")
		}

		return nil
	})

	return err
}

func (d *Db) PutPackage(name string, pkg *action.Manifest) error {
	db, err := d.getDb("image")
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("packages"))
		err := b.Put([]byte(name), []byte(pkg.Json()))
		return err
	})

	return err
}