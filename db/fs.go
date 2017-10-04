package db

import (
	"bytes"
	"errors"
	"strings"
	//"fmt"
	"github.com/boltdb/bolt"
)

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

func GetFsEntry(path string) (*FsEntry, error) {
	db, err := GetDb("image")
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
				fsEntry.Pkgs[fsPkg(string(k))] = true
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

func DelFsEntry(path string, pkg string) error {
	db, err := GetDb("image")
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
		v := b.Delete([]byte(fsKey(path, pkg)))

		if v != nil {
			return errors.New("FS object not recorded")
		}

		return nil
	})

	return err
}

func PutFsEntry(path string, pkg string, typ string) error {
	db, err := GetDb("image")
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
		err := b.Put([]byte(fsKey(path, pkg)), []byte(typ))
		return err
	})

	return err
}

func fsKey(path string, pkg string) string {
	return strings.Join([]string{path, pkg}, "\x00")
}

func fsPkg(key string) string {
	return strings.Split(key, "\x00")[1]
}
