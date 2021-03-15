package db

import (
	"enroller/pkg/scep/crypto"
)

type DBSCEPStore interface {
	InsertCRT(crypto.CRT) error
	SelectCRT(dn string, serial string) (crypto.CRT, error)
	GetCRTs() (crypto.CRTs, error)
	RevokeCRT(dn string, serial string) error
	Delete(dn string, serial string) error
}
