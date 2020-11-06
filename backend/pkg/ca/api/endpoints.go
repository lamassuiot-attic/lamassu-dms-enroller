package api

import (
	"context"
	"enroller/pkg/ca/secrets"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	GetCAsEndpoint    endpoint.Endpoint
	GetCAInfoEndpoint endpoint.Endpoint
	DeleteCAEndpoint  endpoint.Endpoint
}

func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		GetCAsEndpoint:    MakeGetCAsEndpoint(s),
		GetCAInfoEndpoint: MakeGetCAInfoEndpoint(s),
		DeleteCAEndpoint:  MakeDeleteCAEndpoint(s),
	}
}

func MakeGetCAsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(getCAsRequest)
		CAs, err := s.GetCAs(ctx)
		return getCAsResponse{CAs: CAs, Err: err}, nil
	}
}

func MakeGetCAInfoEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getCAInfoRequest)
		CAInfo, err := s.GetCAInfo(ctx, req.CA)
		return getCAInfoResponse{CAInfo: CAInfo, Err: err}, nil
	}
}

func MakeDeleteCAEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteCARequest)
		err = s.DeleteCA(ctx, req.CA)
		return deleteCAResponse{Err: err}, nil
	}
}

type getCAsRequest struct{}

type getCAsResponse struct {
	CAs secrets.CAs
	Err error
}

func (r getCAsResponse) error() error { return r.Err }

type getCAInfoRequest struct {
	CA string
}

type getCAInfoResponse struct {
	CAInfo secrets.CAInfo
	Err    error
}

func (r getCAInfoResponse) error() error { return r.Err }

type deleteCARequest struct {
	CA string
}

type deleteCAResponse struct {
	Err error
}

func (r deleteCAResponse) error() error { return r.Err }
