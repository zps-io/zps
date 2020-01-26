package zpm

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/fezz-io/zps/action"
)

type Security interface {
	Verify(publisher string, content *[]byte, signatures []*action.Signature) error
}

func NewSecurity(mode string, pki *Pki) (Security, error) {
	// Setup initial PKI verify opts, loading CAs and Intermediates from pki store
	cas := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	caEntries, err := pki.Certificates.GetByType(PKICertCA)
	if err != nil {
		return nil, err
	}

	intEntries, err := pki.Certificates.GetByType(PKICertIntermediate)
	if err != nil {
		return nil, err
	}

	for _, crt := range caEntries {
		ok := cas.AppendCertsFromPEM(crt.Cert)
		if !ok {
			return nil, fmt.Errorf("%s loaded from pki db did not parse", crt.Fingerprint)
		}
	}

	for _, crt := range intEntries {
		ok := cas.AppendCertsFromPEM(crt.Cert)
		if !ok {
			return nil, fmt.Errorf("%s loaded from pki db did not parse", crt.Fingerprint)
		}
	}

	switch mode {
	case SecurityModeOffline:
		return &SecurityOffline{pki, cas, intermediates}, nil
	default:
		return nil, errors.New("security mode does not exist")
	}
}

func SecurityValidateBytes(content *[]byte, cert *x509.Certificate, signature action.Signature) error {
	switch signature.Algo {
	case "sha256":
		sig, _ := hex.DecodeString(signature.Value)
		hash := sha256.Sum256(*content)

		err := rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), crypto.SHA256, hash[:], sig)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported signature algorithm")
	}

	return nil
}
