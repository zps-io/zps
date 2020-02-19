/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zpm

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"path/filepath"
	"time"

	"github.com/asdine/storm"
	bolt "go.etcd.io/bbolt"
)

const (
	PKICertCA           = "ca"
	PKICertIntermediate = "intermediate"
	PKICertUser         = "user"
	PKICertCRL          = "crl"
)

type Pki struct {
	Path         string
	Certificates *PkiCertificates
	KeyPairs     *PkiKeyPairs
}

type PkiCertificates struct {
	getDb func() (*storm.DB, error)
}

type PkiKeyPairs struct {
	getDb func() (*storm.DB, error)
}

type CertEntry struct {
	Fingerprint string `storm:"id"`
	Publisher   string `storm:"index"`
	Type        string `storm:"index"`
	Cert        []byte
}

type KeyPairEntry struct {
	Fingerprint string `storm:"id"`
	Subject     string `storm:"index"`
	Publisher   string `storm:"index"`
	Cert        []byte
	Key         []byte
}

func NewPki(path string) *Pki {
	pki := &Pki{Path: path}
	pki.Certificates = &PkiCertificates{}
	pki.Certificates.getDb = pki.getDb

	pki.KeyPairs = &PkiKeyPairs{}
	pki.KeyPairs.getDb = pki.getDb

	return pki
}

func NewCertEntry(fingerprint string, publisher string, typ string, cert []byte) *CertEntry {
	return &CertEntry{Fingerprint: fingerprint, Publisher: publisher, Type: typ, Cert: cert}
}

func (p *Pki) getDb() (*storm.DB, error) {
	db, err := storm.Open(filepath.Join(p.Path, "pki.db"), storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second}))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (p *PkiCertificates) All() ([]*CertEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*CertEntry

	err = db.All(&entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiCertificates) Get(fingerprint string) (*CertEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entry CertEntry

	err = db.One("Fingerprint", fingerprint, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &entry, err
}

func (p *PkiCertificates) GetByPublisher(publisher string) ([]*CertEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*CertEntry

	err = db.Find("Publisher", publisher, &entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiCertificates) GetByType(typ string) ([]*CertEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*CertEntry

	err = db.Find("Type", typ, &entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiCertificates) Del(fingerprint string) error {
	db, err := p.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.DeleteStruct(&CertEntry{Fingerprint: fingerprint})

	return err
}

func (p *PkiCertificates) Put(fingerprint string, publisher string, typ string, cert []byte) error {
	db, err := p.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	entry := &CertEntry{fingerprint, publisher, typ, cert}

	err = db.Save(entry)
	return err
}

func (p *PkiKeyPairs) All() ([]*KeyPairEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*KeyPairEntry

	err = db.All(&entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiKeyPairs) Get(fingerprint string) (*KeyPairEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entry KeyPairEntry

	err = db.One("Fingerprint", fingerprint, &entry)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &entry, err
}

func (p *PkiKeyPairs) GetByPublisher(publisher string) ([]*KeyPairEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*KeyPairEntry

	err = db.Find("Publisher", publisher, &entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiKeyPairs) GetBySubject(subject string) ([]*KeyPairEntry, error) {
	db, err := p.getDb()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var entries []*KeyPairEntry

	err = db.Find("Subject", subject, &entries)
	if err != nil {
		if err == storm.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	return entries, err
}

func (p *PkiKeyPairs) Del(fingerprint string) error {
	db, err := p.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.DeleteStruct(&KeyPairEntry{Fingerprint: fingerprint})

	return err
}

func (p *PkiKeyPairs) Put(fingerprint string, subject string, publisher string, cert []byte, key []byte) error {
	db, err := p.getDb()
	if err != nil {
		return err
	}
	defer db.Close()

	entry := &KeyPairEntry{fingerprint, subject, publisher, cert, key}

	err = db.Save(entry)
	return err
}

func (k *KeyPairEntry) RSAKey() (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(k.Key)
	if block == nil {
		return nil, errors.New("failed to decode key")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
