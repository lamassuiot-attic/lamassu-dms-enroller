package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"time"
)

type CRT struct {
	Status         string `json:"status"`
	ExpirationDate string `json:"expirationDate"`
	RevocationDate string `json:"revocationDate"`
	Serial         int64  `json:"serial"`
	DN             string `json:"dn"`
}

type CRTs struct {
	CRTs []CRT `json:"-"`
}

type CSR struct {
	Id                     int    `json:"id"`
	CountryName            string `json:"c"`
	StateOrProvinceName    string `json:"st"`
	LocalityName           string `json:"l"`
	OrganizationName       string `json:"o"`
	OrganizationalUnitName string `json:"ou,omitempty"`
	CommonName             string `json:"cn"`
	EmailAddress           string `json:"mail,omitempty"`
	Status                 string `json:"status"`
	CsrFilePath            string `json:"csrpath,omitempty"`
}

type CSRs struct {
	CSRs []CSR `json:"-"`
}

const (
	csrPEMBlockType = "CERTIFICATE REQUEST"
	PublicKeyHeader = "-----BEGIN PUBLIC KEY-----"
	PublicKeyFooter = "-----END PUBLIC KEY-----"
	caPEMBlockType  = "CERTIFICATE"
	PendingStatus   = "NEW"
	ApprobedStatus  = "APPROBED"
	DeniedStatus    = "DENIED"
)

func ParseKeycloakPublicKey(data []byte) (*rsa.PublicKey, error) {
	pubPem, _ := pem.Decode(data)
	parsedKey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
		return nil, errors.New("Unable to parse public key")
	}
	pubKey := parsedKey.(*rsa.PublicKey)
	return pubKey, nil
}

func ParseNewCSR(data []byte) (CSR, error) {
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return CSR{}, errors.New("Cannot find the next PEM formatted block")
	}
	if pemBlock.Type != csrPEMBlockType || len(pemBlock.Headers) != 0 {
		return CSR{}, errors.New("Unmatched type of headers")
	}
	certReq, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		return CSR{}, err
	}
	csr := CSR{
		CountryName:            strings.Join(certReq.Subject.Country, " "),
		StateOrProvinceName:    strings.Join(certReq.Subject.Province, " "),
		LocalityName:           strings.Join(certReq.Subject.Locality, " "),
		OrganizationName:       strings.Join(certReq.Subject.Organization, " "),
		OrganizationalUnitName: strings.Join(certReq.Subject.OrganizationalUnit, " "),
		EmailAddress:           strings.Join(certReq.EmailAddresses, " "),
		CommonName:             certReq.Subject.CommonName,
		Status:                 PendingStatus,
	}
	return csr, nil
}

func (csr CSR) SignCSR(csrData []byte, serial *big.Int, caCertData []byte, caKeyData []byte) ([]byte, error) {
	pemBlock, _ := pem.Decode(csrData)
	if pemBlock == nil {
		return nil, errors.New("Cannot find the next PEM formatted block")
	}
	if pemBlock.Type != csrPEMBlockType || len(pemBlock.Headers) != 0 {
		return nil, errors.New("Unmatched type of headers")
	}
	certReq, err := x509.ParseCertificateRequest(pemBlock.Bytes)

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      certReq.Subject,
		NotBefore:    time.Now().Add(-600).UTC(),
		NotAfter:     time.Now().AddDate(0, 0, 365).UTC(),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
		},
		SignatureAlgorithm: certReq.SignatureAlgorithm,
	}
	caCert, err := parseCACert(caCertData)
	if err != nil {
		return nil, err
	}
	caKey, err := parseCAKey(caKeyData)
	if err != nil {
		return nil, err
	}
	crtBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, certReq.PublicKey, caKey)

	if err != nil {
		return nil, errors.New("Unable to sign the certificate signing request")
	}
	return crtBytes, nil

}

func parseCACert(data []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(data)
	if pemBlock == nil {
		return nil, errors.New("Cannot find the next PEM formatted block")
	}
	if pemBlock.Type != caPEMBlockType || len(pemBlock.Headers) != 0 {
		return nil, errors.New("Unmatched type of headers")
	}
	caCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, errors.New("Unable to parse CA certificate")
	}
	return caCert, nil
}

func parseCAKey(data []byte) (*rsa.PrivateKey, error) {
	pivPem, _ := pem.Decode(data)
	pivKey, err := x509.ParsePKCS1PrivateKey(pivPem.Bytes)
	if err != nil {
		return nil, errors.New("Unable to parse CA key")
	}
	return pivKey, nil
}
