package api

import (
	"context"
	"enroller/pkg/ca/secrets"
	"sync"
)

type Service interface {
	GetCAs(ctx context.Context) (secrets.CAs, error)
	GetCAInfo(ctx context.Context, CA string) (secrets.CAInfo, error)
	DeleteCA(ctx context.Context, CA string) error
}

type caService struct {
	mtx     sync.RWMutex
	secrets secrets.Secrets
}

func NewCAService(secrets secrets.Secrets) Service {
	return &caService{
		secrets: secrets,
	}
}

func (s *caService) GetCAs(ctx context.Context) (secrets.CAs, error) {
	CAs, err := s.secrets.GetCAs()
	if err != nil {
		return secrets.CAs{}, err
	}
	return CAs, nil

}

func (s *caService) GetCAInfo(ctx context.Context, CA string) (secrets.CAInfo, error) {
	CAInfo, err := s.secrets.GetCAInfo(CA)
	if err != nil {
		return secrets.CAInfo{}, err
	}
	return CAInfo, nil

}

func (s *caService) DeleteCA(ctx context.Context, CA string) error {
	err := s.secrets.DeleteCA(CA)
	if err != nil {
		return err
	}
	return nil
}
