package secrets

import "crypto/x509"

type Secrets interface {
	SignCSR(csr *x509.CertificateRequest) ([]byte, error)
}
