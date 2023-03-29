package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	CERT_ORG = "Portainer.io"
	CERT_C   = "NZ"
	CERT_L   = "Auckland"
)

// GenerateSelfSignedTLSCertificates generates a self-signed TLS certificate and key
func GenerateSelfSignedTLSCertificates(logger *zap.SugaredLogger, keyPath, certificatePath string) error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{CERT_ORG},
			Country:      []string{CERT_C},
			Locality:     []string{CERT_L},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logger.Errorf("Unable to generate CA private key: %s", err)
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		logger.Errorf("Unable to generate CA TLS certificate: %s", err)
		return err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{CERT_ORG},
			Country:      []string{CERT_C},
			Locality:     []string{CERT_L},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		logger.Errorf("Unable to generate TLS certificate private key: %s", err)
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		logger.Errorf("Unable to generate TLS certificate: %s", err)
		return err
	}

	certOut, err := os.Create(certificatePath)
	if err != nil {
		logger.Errorf("Unable to open %s for writing: %s", certificatePath, err)
		return err
	}

	pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	err = certOut.Close()
	if err != nil {
		logger.Errorf("Error closing %s: %s", certificatePath, err)
		return err
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logger.Errorf("Failed to open %s for writing: %v", keyPath, err)
		return err
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(certPrivKey)
	if err != nil {
		logger.Errorf("Unable to marshal private key: %s", err)
		return err
	}

	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	err = keyOut.Close()
	if err != nil {
		logger.Errorf("Error closing %s: %s", keyPath, err)
		return err
	}

	logger.Info("TLS certificates generated")
	return nil
}
