package files

import "math/big"

type FileCSRStore interface {
	InsertFileCSR(id int, rawData []byte) error
	InsertFileCert(id int, rawData []byte) error
	SelectFileByID(id int) ([]byte, error)
	DeleteFile(id int) error
	Serial() (*big.Int, error)
	writeSerial(serial *big.Int) error
	LoadCACert() ([]byte, error)
	LoadCAKey() ([]byte, error)
	LoadCert(id int) ([]byte, error)
}
