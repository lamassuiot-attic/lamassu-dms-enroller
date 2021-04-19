package store

import (
	"math/big"

	"github.com/lamassuiot/enroller/pkg/enroller/models/certs"
)

type DB interface {
	Insert(crt certs.CRT) error
	Serial() (*big.Int, error)
	Revoke(id int, revocationDate string) error
	Delete(id int) error
}

type File interface {
	Insert(id int, data []byte) error
	SelectByID(id int) ([]byte, error)
	Delete(id int) error
}
