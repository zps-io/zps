package db

import (
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"

	"github.com/solvent-io/zps/zps"
)

func GetDb(name string) (*bolt.DB, error) {
	path, err := zps.DbPath()
	if err != nil {
		return nil, err
	}

	db, err := bolt.Open(filepath.Join(path, name+".db"), 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		return nil, err
	}

	return db, nil
}
