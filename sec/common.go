package sec

import (
	"bytes"
	"crypto"
	"crypto/rand"
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

	if len(cert.Subject.Organization) == 0 {
		return "", "", "", errors.New("invalid certificate organization")
	}

	return cert.Subject.CommonName, cert.Subject.Organization[0], fingerprint, nil
}

func SecuritySignBytes(content *[]byte, certFingerprint string, key *rsa.PrivateKey, algo string) (*action.Signature, error) {
	switch algo {
	case "sha256":
		digest := sha256.Sum256(*content)

		rng := rand.Reader

		signature, err := rsa.SignPKCS1v15(rng, key, crypto.SHA256, digest[:])
		if err != nil {
			return nil, err
		}

		return &action.Signature{
			FingerPrint: certFingerprint,
			Algo:        "sha256",
			Value:       hex.EncodeToString(signature),
		}, nil

	default:
		return nil, errors.New("unsupported signature algorithm")
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
