package tlscert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type CertPaths struct {
	CertPath string
	KeyPath  string
}

func ProvideCertAndKey() (paths CertPaths, err error) {
	certDir := "cert"
	certPath := filepath.Join(certDir, "cert.pem")
	keyPath := filepath.Join(certDir, "key.pem")

	if _, statErr := os.Stat(certPath); statErr == nil {
		if _, statErr := os.Stat(keyPath); statErr == nil {
			paths.CertPath = certPath
			paths.KeyPath = keyPath
			return
		}
	}

	if err = os.MkdirAll(certDir, 0755); err != nil {
		return
	}

	defer func() {
		if err != nil {
			os.Remove(certPath)
			os.Remove(keyPath)
		}
	}()

	privateKey, keyErr := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if keyErr != nil {
		err = keyErr
		return
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

	certDER, certErr := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if certErr != nil {
		err = certErr
		return
	}

	certFile, ferr := os.Create(certPath)
	if ferr != nil {
		err = ferr
		return
	}
	if err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		certFile.Close()
		return
	}
	if err = certFile.Close(); err != nil {
		return
	}

	keyFile, ferr := os.Create(keyPath)
	if ferr != nil {
		err = ferr
		return
	}
	keyDER, kerr := x509.MarshalECPrivateKey(privateKey)
	if kerr != nil {
		err = kerr
		keyFile.Close()
		return
	}
	if err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		keyFile.Close()
		return
	}
	if err = keyFile.Close(); err != nil {
		return
	}

	paths.CertPath = certPath
	paths.KeyPath = keyPath
	return
}
