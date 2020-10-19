package db

import (
	"enroller/pkg/enroller/crypto"
)

type DBCSRStore interface {
	InsertCSR(csr crypto.CSR) (int, error)
	SelectCSRsByStatus(status string) crypto.CSRs
	SelectCSRByID(id int) (crypto.CSR, error)
	UpdateCSRByID(id int, csr crypto.CSR) (crypto.CSR, error)
	UpdateCSRFilePath(csr crypto.CSR) error
}
