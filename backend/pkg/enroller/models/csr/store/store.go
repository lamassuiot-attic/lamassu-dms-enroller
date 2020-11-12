package store

import "enroller/pkg/enroller/models/csr"

type DB interface {
	Insert(c csr.CSR) (int, error)
	SelectAll() csr.CSRs
	SelectAllByCN(cn string) csr.CSRs
	SelectByStatus(status string) csr.CSRs
	SelectByID(id int) (csr.CSR, error)
	UpdateByID(id int, c csr.CSR) (csr.CSR, error)
	UpdateFilePath(c csr.CSR) error
}

type File interface {
	Insert(id int, data []byte) error
	SelectByID(id int) ([]byte, error)
	Delete(id int) error
}
