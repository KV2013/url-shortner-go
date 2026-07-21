package tlscert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertPaths содержит пути к TLS сертификату и ключу.
type CertPaths struct {
	CertPath string
	KeyPath  string
}

// ProvideCertAndKey генерирует TLS сертификат и ключ, если они не существуют.
func ProvideCertAndKey() (CertPaths, error) {
	certDir := "cert"
	certPath := filepath.Join(certDir, "cert.pem")
	keyPath := filepath.Join(certDir, "key.pem")

	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return CertPaths{CertPath: certPath, KeyPath: keyPath}, nil
		}
	}

	if err := os.MkdirAll(certDir, 0755); err != nil {
		return CertPaths{}, err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return CertPaths{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return CertPaths{}, err
	}

	certFile, err := os.Create(certPath)
	if err != nil {
		return CertPaths{}, err
	}
	defer certFile.Close()
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return CertPaths{}, err
	}

	keyFile, err := os.Create(keyPath)
	if err != nil {
		return CertPaths{}, err
	}
	defer keyFile.Close()
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}); err != nil {
		return CertPaths{}, err
	}

	return CertPaths{CertPath: certPath, KeyPath: keyPath}, nil
}
