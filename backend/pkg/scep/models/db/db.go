package db

import (
	"enroller/pkg/scep/crypto"
)

type DBSCEPStore interface {
	GetCRTs() (crypto.CRTs, error)
	RevokeCRT(dn string, serial string) error
}
