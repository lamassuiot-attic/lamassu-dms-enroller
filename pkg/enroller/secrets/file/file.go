package file

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"enroller/pkg/enroller/crypto"
	certstore "enroller/pkg/enroller/models/certs/store"
	"enroller/pkg/enroller/secrets"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type File struct {
	CACert       string
	CAKey        string
	OCSPServer   string
	certsDBStore certstore.DB
	logger       log.Logger
}

func NewFile(CACert string, CAKey string, OCSPServer string, certsDBStore certstore.DB, logger log.Logger) secrets.Secrets {
	return &File{CACert: CACert, CAKey: CAKey, OCSPServer: OCSPServer, certsDBStore: certsDBStore, logger: logger}
}

func (f *File) SignCSR(csr *x509.CertificateRequest) ([]byte, error) {
	caCert, err := loadCACert(f.CACert)
	if err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Could not load CA certificate")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "CA certificate loaded")
	caKey, err := loadCAKey(f.CAKey)
	if err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Could not load CA key")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "CA key loaded")
	serial, err := f.certsDBStore.Serial()
	if err != nil {
		level.Error(f.logger).Log("err", err, "msg", "Could not get serial from database")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "Serial obtained from database")
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      csr.Subject,
		NotBefore:    time.Now().Add(-600).UTC(),
		NotAfter:     time.Now().AddDate(0, 0, 365).UTC(),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		OCSPServer:   []string{f.OCSPServer},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		SignatureAlgorithm: csr.SignatureAlgorithm,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, caCert, csr.PublicKey, caKey)
	if err != nil {

		f.logger.Log("err", err, "msg", "Could not create signed certificate")
		return nil, err
	}
	level.Info(f.logger).Log("msg", "CSR with serial "+fmt.Sprintf("%x", serial)+" signed by Enroller CA")
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
