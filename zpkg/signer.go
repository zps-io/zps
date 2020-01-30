package zpkg

import (
	"bufio"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"

	"github.com/fezz-io/zps/sec"
	"github.com/fezz-io/zps/zpkg/payload"
)

type Signer struct {
	reader   *Reader
	workPath string
}

func NewSigner(path string, workPath string) *Signer {
	signer := &Signer{}
	signer.reader = NewReader(path, workPath)
	signer.workPath = workPath

	return signer
}

func (s *Signer) Sign(fingerprint string, key *[]byte) error {
	err := s.reader.Read()
	if err != nil {
		return err
	}

	block, _ := pem.Decode(*key)
	if block == nil {
		return errors.New("failed to decode key")
	}

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	var content []byte
	content = []byte(s.reader.Manifest.ToSigningJson())

	// Get signature action
	sigAction, err := sec.SecuritySignBytes(&content, fingerprint, rsaKey, "sha256")

	// Modify manifest
	manifest := s.reader.Manifest
	manifest.Add(sigAction)

	payloadOffset := s.reader.Payload.Offset()
	s.reader.Close()

	// Write new header and manifest
	writer := NewWriter()
	tmpFile := s.reader.path + ".signing"

	err = writer.Write(tmpFile, NewHeader(Version, Compression), manifest, payload.NewWriter("", 0))
	if err != nil {
		return err
	}

	// Copy payload to temp file

	source, err := os.Open(s.reader.path)
	if err != nil {
		return err
	}
	_, err = source.Seek(payloadOffset, 0)
	if err != nil {
		return err
	}

	srcReader := bufio.NewReader(source)

	dest, err := os.OpenFile(tmpFile, os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}

	_, err = dest.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	destWriter := bufio.NewWriter(dest)

	_, err = io.Copy(destWriter, srcReader)
	if err != nil {
		return err
	}

	destWriter.Flush()
	source.Close()
	dest.Close()

	// TODO Probably doesn't work cross mount

	os.Remove(s.reader.path)
	os.Rename(tmpFile, s.reader.path)

	return nil
}
