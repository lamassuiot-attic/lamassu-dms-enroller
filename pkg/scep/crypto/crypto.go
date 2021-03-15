package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
)

type CRT struct {
	Status         string `json:"status"`
	ExpirationDate string `json:"expirationDate"`
	RevocationDate string `json:"revocationDate"`
	Serial         string `json:"serial"`
	DN             string `json:"dn"`
	CRTPath        string `json:"crtpath"`
	Key            string `json:"key"`
	KeySize        int    `json:"keySize"`
}

type CRTs struct {
	CRTs []CRT `json:""`
}

const (
	csrPEMBlockType = "CERTIFICATE REQUEST"
	PublicKeyHeader = "-----BEGIN PUBLIC KEY-----"
	PublicKeyFooter = "-----END PUBLIC KEY-----"
	caPEMBlockType  = "CERTIFICATE"
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

func CreateCAPool(CAPath string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(CAPath)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}
