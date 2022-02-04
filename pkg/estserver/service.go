package estserver

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/log/level"
	"github.com/lamassuiot/dms-enroller/pkg/utils"
	lamassuca "github.com/lamassuiot/lamassu-ca/client"
	lamassuest "github.com/lamassuiot/lamassu-est/pkg/server/api"
)

var (
	oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}
)

type EstService struct {
	mtx             sync.RWMutex
	logger          log.Logger
	lamassuCaClient lamassuca.LamassuCaClient
}

func NewEstService(lamassuCaClient *lamassuca.LamassuCaClient, logger log.Logger) lamassuest.Service {
	return &EstService{
		lamassuCaClient: *lamassuCaClient,
		logger:          logger,
	}
}

func (s *EstService) Health(ctx context.Context) bool {
	return true
}

func (s *EstService) CACerts(ctx context.Context, aps string, r *http.Request) ([]*x509.Certificate, error) {
	certs, err := s.lamassuCaClient.GetCAs(ctx, "dmsenroller")
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Error in client request")
		return nil, err
	}

	x509Certificates := []*x509.Certificate{}
	for _, v := range certs.Certs {
		data, _ := base64.StdEncoding.DecodeString(v.CertContent.CerificateBase64)
		block, _ := pem.Decode([]byte(data))
		cert, _ := x509.ParseCertificate(block.Bytes)
		x509Certificates = append(x509Certificates, cert)
	}

	return x509Certificates, nil
}

func (s *EstService) Enroll(ctx context.Context, csr *x509.CertificateRequest, aps string, crt *x509.Certificate, r *http.Request) (*x509.Certificate, error) {
	dataCert, err := s.lamassuCaClient.SignCertificateRequest(ctx, aps, csr, "dmsenroller")
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Error in client request")
		return &x509.Certificate{}, err
	}
	return dataCert, nil
}

func (s *EstService) Reenroll(ctx context.Context, cert *x509.Certificate, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, error) {
	// Compare Subject fields
	if !reflect.DeepEqual(cert.Subject, csr.Subject) {
		return nil, lamassuest.ErrSubjectChanged
	}

	// Compare SubjectAltName fields
	var csrSAN pkix.Extension
	var certSAN pkix.Extension

	for _, ext := range csr.Extensions {
		if ext.Id.Equal(oidSubjectAltName) {
			csrSAN = ext
			break
		}
	}
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(oidSubjectAltName) {
			certSAN = ext
			break
		}
	}

	if !bytes.Equal(csrSAN.Value, certSAN.Value) {
		return nil, lamassuest.ErrSubjectChanged

	}
	dataCert, err := s.lamassuCaClient.SignCertificateRequest(ctx, aps, csr, "dmsenroller")
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Error in client request")
		return &x509.Certificate{}, err
	}
	return dataCert, nil
}
func (s *EstService) ServerKeyGen(ctx context.Context, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, []byte, error) {
	csrkey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate temporary private key: %v", err)
	}
	privkey, err := x509.MarshalPKCS8PrivateKey(csrkey)
	if err != nil {
		return nil, nil, err
	}
	csr, err = utils.GenerateCSR(csr, csrkey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate new csr: %v", err)
	}

	dataCert, err := s.lamassuCaClient.SignCertificateRequest(ctx, aps, csr, "pki")
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Error in client request")
		return &x509.Certificate{}, nil, err
	}
	return dataCert, privkey, nil
}
