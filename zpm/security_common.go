package zpm

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"

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

func SecurityCertMetaFromBytes(certPem *[]byte) (string, string, string, error) {
	block, _ := pem.Decode(*certPem)
	if block == nil {
		return "", "", "", errors.New("failed to parse certificate pem")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", "", "", errors.New("failed to parse certificate: " + err.Error())
	}

	fingerprint := SpkiFingerprint(cert).String()

	return cert.Subject.CommonName, cert.DNSNames[0], fingerprint, nil
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

func SecurityValidateKeyPair(certPath string, keyPath string) error {
	_, err := tls.LoadX509KeyPair(certPath, keyPath)

	return err
}

type Fingerprint []byte

func (f Fingerprint) String() string {
	var buf bytes.Buffer
	for i, b := range f {
		if i > 0 {
			fmt.Fprintf(&buf, ":")
		}
		fmt.Fprintf(&buf, "%02x", b)
	}
	return buf.String()
}

func ParseFingerprint(fp string) (Fingerprint, error) {
	s := strings.Join(strings.Split(fp, ":"), "")
	buf, err := hex.DecodeString(s)
	return Fingerprint(buf), err
}

func SpkiFingerprint(cert *x509.Certificate) Fingerprint {
	h := sha256.New()
	h.Write(cert.RawSubjectPublicKeyInfo)
	return Fingerprint(h.Sum(nil))
}
