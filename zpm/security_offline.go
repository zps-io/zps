package zpm

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

	"github.com/fezz-io/zps/sec"

	"github.com/fezz-io/zps/action"
)

const (
	SecurityModeOffline = "offline"
)

type SecurityOffline struct {
	pki *Pki

	caCache           *x509.CertPool
	intermediateCache *x509.CertPool
}

func (s *SecurityOffline) Mode() string {
	return SecurityModeOffline
}

func (s *SecurityOffline) KeyPair(publisher string) (*KeyPairEntry, error) {
	pairs, err := s.pki.KeyPairs.GetByPublisher(publisher)
	if err != nil {
		return nil, err
	}

	if len(pairs) > 0 {
		return pairs[0], nil
	}

	return nil, nil
}

func (s *SecurityOffline) Trust(content *[]byte, typ string) (string, string, error) {
	subject, publisher, fingerprint, err := sec.SecurityCertMetaFromBytes(content)
	if err != nil {
		return "", "", err
	}

	// Attempt to detect cert type
	if typ == "" {
		if strings.Contains(subject, "CA") {
			typ = PKICertCA
		} else if strings.Contains(subject, "Intermediate") {
			typ = PKICertIntermediate
		} else {
			typ = PKICertUser
		}
	}

	err = s.pki.Certificates.Put(fingerprint, subject, publisher, typ, *content)
	if err != nil {
		return "", "", err
	}

	return subject, publisher, nil
}

// TODO warn on the presence of invalid signatures
func (s *SecurityOffline) Verify(content *[]byte, signatures []*action.Signature) (*action.Signature, error) {
	if len(signatures) == 0 {
		return nil, errors.New("no signatures present")
	}

	// Setup verify opts
	opts := x509.VerifyOptions{
		Roots:         s.caCache,
		Intermediates: s.intermediateCache,
	}

	for _, sig := range signatures {
		// Load cert if found
		certEntry, err := s.pki.Certificates.Get(sig.FingerPrint)
		if err != nil || certEntry == nil {
			continue
		}

		asn, _ := pem.Decode(certEntry.Cert)
		if asn == nil {
			return nil, fmt.Errorf("failed to parse pem for cert entry: %s", certEntry.Fingerprint)
		}

		cert, err := x509.ParseCertificate(asn.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse asn for cert entry: %s", certEntry.Fingerprint)
		}

		// TODO Check CRL, if found

		// TODO for now return on first successful validation
		if s.validateChain(opts, cert) == nil && sec.SecurityValidateBytes(content, cert, *sig) == nil {
			return sig, nil
		}
	}

	return nil, errors.New("no trusted certificates found for signatures")
}

func (s *SecurityOffline) validateChain(opts x509.VerifyOptions, certificate *x509.Certificate) error {
	_, err := certificate.Verify(opts)

	return err
}
