package file

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"enroller/pkg/enroller/crypto"
	certstore "enroller/pkg/enroller/models/certs/store"
	"enroller/pkg/enroller/secrets"
	"io/ioutil"
	"time"
)

type File struct {
	CACert       string
	CAKey        string
	certsDBStore certstore.DB
}

func NewFile(CACert string, CAKey string, certsDBStore certstore.DB) secrets.Secrets {
	return &File{CACert: CACert, CAKey: CAKey, certsDBStore: certsDBStore}
}

func (f *File) SignCSR(csr *x509.CertificateRequest) ([]byte, error) {
	caCert, err := loadCACert(f.CACert)
	if err != nil {
		return nil, err
	}
	caKey, err := loadCAKey(f.CAKey)
	if err != nil {
		return nil, err
	}
	serial, err := f.certsDBStore.Serial()
	if err != nil {
		return nil, err
	}
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      csr.Subject,
		NotBefore:    time.Now().Add(-600).UTC(),
		NotAfter:     time.Now().AddDate(0, 0, 365).UTC(),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		SignatureAlgorithm: csr.SignatureAlgorithm,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, caCert, csr.PublicKey, caKey)
	if err != nil {
		return nil, err
	}
	return cert, nil

}

func loadCACert(CACert string) (*x509.Certificate, error) {
	certPEM, err := ioutil.ReadFile(CACert)
	if err != nil {
		return nil, err
	}
	pemBlock, _ := pem.Decode(certPEM)
	err = crypto.CheckPEMBlock(pemBlock, crypto.CertPEMBlockType)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func loadCAKey(CAKey string) (*rsa.PrivateKey, error) {
	keyPEM, err := ioutil.ReadFile(CAKey)
	if err != nil {
		return nil, err
	}
	pemBlock, _ := pem.Decode(keyPEM)
	err = crypto.CheckPEMBlock(pemBlock, crypto.KeyPEMBlockType)
	if err != nil {
		return nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}
