package zpm

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/coreos/bbolt"
	"github.com/solvent-io/zps/action"
)

type State struct {
	Path         string
	Packages     *StatePackages
	Objects      *StateObjects
	Transactions *StateTransactions
}

type StatePackages struct {
	getDb func() (*storm.DB, error)
}

type StateObjects struct {
	getDb func() (*storm.DB, error)
}

type StateTransactions struct {
	getDb func() (*storm.DB, error)
}

type PkgEntry struct {
	Name     string `storm:"id"`
	Manifest []byte
}

type FsEntry struct {
	Key  string `storm:"id"`
	Path string `storm:"index"`
	Pkg  string `storm:"index"`
	Type string `storm:"index"`
}

type TransactionEntry struct {
	Key       string `storm:"id"`
	Id        string `storm:"index"`
	PkgId     string
	Operation string
	Date      *time.Time `storm:"index"`
}

func NewState(path string) *State {
	state := &State{Path: path}
	state.Packages = &StatePackages{}
	state.Packages.getDb = state.getDb

	state.Objects = &StateObjects{}
	state.Objects.getDb = state.getDb

	state.Transactions = &StateTransactions{}
	state.Transactions.getDb = state.getDb

	return state
}

func NewFsEntry(path string, pkg string, typ string) *FsEntry {
	fs := &FsEntry{Path: path, Pkg: pkg, Type: typ}
	fs.Key = strings.Join([]string{path, pkg}, "\x00")

	return fs
}

func NewTransactionEntry(id string, pkgId string, operation string, date *time.Time) *TransactionEntry {
	ts := &TransactionEntry{Id: id, PkgId: pkgId, Operation: operation, Date: date}
	ts.Key = strings.Join([]string{id, pkgId}, "\x00")

	return ts
}

func (s *State) getDb() (*storm.DB, error) {
	db, err := storm.Open(filepath.Join(s.Path, "image.db"), storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *StatePackages) All() ([]*action.Manifest, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*PkgEntry
	var packages []*action.Manifest

	err = db.All(&entries)

	for _, pkg := range entries {
		manifest := action.NewManifest()
		err := manifest.Load(string(pkg.Manifest))
		if err != nil {
			return nil, err
		}

		packages = append(packages, manifest)
	}

	return packages, nil
}

func (s *StatePackages) Get(name string) (*action.Manifest, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entry PkgEntry
	pkg := action.NewManifest()

	err = db.One("Name", name, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	err = pkg.Load(string(entry.Manifest))
	if err != nil {
		return nil, err
	}

	return pkg, err
}

func (s *StatePackages) Del(name string) error {
	db, err := s.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.DeleteStruct(&PkgEntry{Name: name})

	return err
}

func (s *StatePackages) Put(name string, pkg *action.Manifest) error {
	db, err := s.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	entry := &PkgEntry{name, []byte(pkg.Json())}

	err = db.Save(entry)
	return err
}

func (s *StateObjects) All() ([]*FsEntry, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*FsEntry

	err = db.All(&entries)

	return entries, nil
}

func (s *StateObjects) Get(path string) ([]*FsEntry, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*FsEntry

	err = db.Find("Path", path, &entries)

	return entries, nil
}

func (s *StateObjects) Del(path string, pkg string) error {
	db, err := s.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.DeleteStruct(NewFsEntry(path, pkg, ""))

	return err
}

func (s *StateObjects) Put(path string, pkg string, typ string) error {
	db, err := s.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Save(NewFsEntry(path, pkg, typ))

	return err
}

// mo moo

func (s *StateTransactions) All() ([]*TransactionEntry, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*TransactionEntry

	err = db.AllByIndex("Date", &entries)

	return entries, nil
}

func (s *StateTransactions) Get(id string) ([]*TransactionEntry, error) {
	db, err := s.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*TransactionEntry

	err = db.Find("Id", id, &entries)

	return entries, nil
}

func (s *StateTransactions) Put(id string, pkgId string, operation string, date *time.Time) error {
	db, err := s.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Save(NewTransactionEntry(id, pkgId, operation, date))

	return err
}
