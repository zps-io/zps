package zpm

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

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

// TODO warn on the presence of invalid signatures
func (s *SecurityOffline) Verify(publisher string, content *[]byte, signatures []*action.Signature) (*action.Signature, error) {
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
