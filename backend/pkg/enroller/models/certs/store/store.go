package store

import (
	"enroller/pkg/enroller/models/certs"
	"math/big"
)

type DB interface {
	Insert(crt certs.CRT) error
	Serial() (*big.Int, error)
	Revoke(id int, revocationDate string) error
}

type File interface {
	Insert(id int, data []byte) error
	SelectByID(id int) ([]byte, error)
}
