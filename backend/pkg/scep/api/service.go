package api

import (
	"context"
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
	ErrIncorrectType    = errors.New("Incorrect media type")
	ErrIncorrectContent = errors.New("Incorrect data content")
	ErrBadRouting       = errors.New("Bad routing")
	ErrBadKey           = errors.New("Unexpected JWT key signing method")
)

func NewSCEPService(scepDB db.DBSCEPStore) Service {
	return &scepService{
		scepDB: scepDB,
	}
}

func (s *scepService) GetSCEPCRTs(ctx context.Context) (crypto.CRTs, error) {
	crts, err := s.scepDB.GetCRTs()
	if err != nil {
		return crypto.CRTs{CRTs: []crypto.CRT{}}, err
	}
	return crts, nil
}

func (s *scepService) RevokeSCEPCRT(ctx context.Context, dn string, serial string) error {
	err := s.scepDB.RevokeCRT(dn, serial)
	return err
}
