package api

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"enroller/pkg/scep/crypto"
	"enroller/pkg/scep/models/db"
)

type Service interface {
	GetSCEPCRTs(ctx context.Context) (crypto.CRTs, error)
	RevokeSCEPCRT(ctx context.Context, dn string, serial string) error
}

type scepService struct {
	mtx    sync.RWMutex
	scepDB db.DBSCEPStore
}

var (
	//Client
	ErrInvalidCert       = errors.New("unable to parse certificate, is invalid")
	ErrInvalidDNOrSerial = errors.New("invalid certificate DN or serial, does not exist")
	ErrInvalidRevokeOp   = errors.New("invalid operation, certificate is already revoked")

	//Server
	ErrGetCertificates = errors.New("unable to get certificates")
	ErrRevokeCert      = errors.New("unable to revoke certificate")
	ErrGetCert         = errors.New("unable to get certificate")
)

func NewSCEPService(scepDB db.DBSCEPStore) Service {
	return &scepService{
		scepDB: scepDB,
	}
}

func (s *scepService) GetSCEPCRTs(ctx context.Context) (crypto.CRTs, error) {
	crts, err := s.scepDB.GetCRTs()
	if err != nil {
		return crypto.CRTs{}, ErrGetCertificates
	}
	return crts, nil
}

func (s *scepService) RevokeSCEPCRT(ctx context.Context, dn string, serial string) error {
	crt, err := s.scepDB.SelectCRT(dn, serial)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidDNOrSerial
		}
		return ErrGetCert
	}
	if crt.Status == "R" {
		return ErrInvalidRevokeOp
	}
	err = s.scepDB.RevokeCRT(dn, serial)
	if err != nil {
		return ErrRevokeCert
	}
	return nil
}
