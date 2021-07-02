package ca

import (
	"context"
	"crypto/x509"
	"github.com/globalsign/est"
	"github.com/lamassuiot/lamassu-est/client/estclient"
	"net/http"
)

type DeviceService struct {

}

func NewVaultService() *DeviceService {
	return &DeviceService{}
}

func (ca *DeviceService) CACerts(ctx context.Context, aps string, req *http.Request) ([]*x509.Certificate, error) {

	var filteredCerts []*x509.Certificate

	return filteredCerts, nil
}

func (ca *DeviceService) Enroll(ctx context.Context, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, error) {
	cert, err := estclient.Enroll(csr, aps)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (ca *DeviceService) CSRAttrs(ctx context.Context, aps string, r *http.Request) (est.CSRAttrs, error) {
	return est.CSRAttrs{}, nil
}

func (ca *DeviceService) Reenroll(ctx context.Context, cert *x509.Certificate, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, error) {
	newCert, err := estclient.Reenroll(csr, aps)
	if err != nil {
		return nil, err
	}
	return newCert, nil
}

func (ca *DeviceService) ServerKeyGen(ctx context.Context, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, []byte, error) {
	return nil, nil, nil
}

func (ca *DeviceService) TPMEnroll(ctx context.Context, csr *x509.CertificateRequest, ekcerts []*x509.Certificate, ekPub, akPub []byte, aps string, r *http.Request) ([]byte, []byte, []byte, error) {
	return nil, nil, nil, nil
}

