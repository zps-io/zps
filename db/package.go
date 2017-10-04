package db

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/solvent-io/zps/action"
)

func Packages() ([]*action.Manifest, error) {
	db, err := GetDb("image")
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

func GetPackage(name string) (*action.Manifest, error) {
	db, err := GetDb("image")
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

func DelPackage(name string) error {
	db, err := GetDb("image")
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

func PutPackage(name string, pkg *action.Manifest) error {
	db, err := GetDb("image")
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
