package api

import (
	"context"

	"github.com/lamassuiot/enroller/pkg/scep/crypto"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint           endpoint.Endpoint
	GetSCEPCRTsEndpoint      endpoint.Endpoint
	PutRevokeSCEPCRTEndpoint endpoint.Endpoint
}

func MakeServerEndpoints(s Service, otTracer stdopentracing.Tracer) Endpoints {
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
		healthEndpoint = opentracing.TraceServer(otTracer, "Health")(healthEndpoint)
	}

	var getSCEPCRTEndpoint endpoint.Endpoint
	{
		getSCEPCRTEndpoint = MakeGetSCEPCRTsEndpoint(s)
		getSCEPCRTEndpoint = opentracing.TraceServer(otTracer, "GetSCEPCRT")(getSCEPCRTEndpoint)
	}
	var putRevokeSCEPCRTEndpoint endpoint.Endpoint
	{
		putRevokeSCEPCRTEndpoint = MakePutRevokeSCEPCRTEndpoint(s)
		putRevokeSCEPCRTEndpoint = opentracing.TraceServer(otTracer, "RevokeSCEPCRT")(putRevokeSCEPCRTEndpoint)
	}
	return Endpoints{
		HealthEndpoint:           healthEndpoint,
		GetSCEPCRTsEndpoint:      getSCEPCRTEndpoint,
		PutRevokeSCEPCRTEndpoint: putRevokeSCEPCRTEndpoint,
	}
}

func MakeHealthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return healthResponse{Healthy: healthy}, nil
	}
}

func MakeGetSCEPCRTsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(getSCEPCRTsRequest)
		CRTs, err := s.GetSCEPCRTs(ctx)
		return getSCEPCRTsResponse{CRTs: CRTs, Err: err}, nil
	}
}

func MakePutRevokeSCEPCRTEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putRevokeSCEPCRTRequest)
		err = s.RevokeSCEPCRT(ctx, req.dn, req.serial)
		return putRevokeSCEPCRTResponse{Err: err}, nil
	}
}

type healthRequest struct{}

type healthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"err,omitempty"`
}

type getSCEPCRTsRequest struct{}

type getSCEPCRTsResponse struct {
	CRTs crypto.CRTs `json:"CRTs,omitempty"`
	Err  error
}

func (r getSCEPCRTsResponse) error() error { return r.Err }

type putRevokeSCEPCRTRequest struct {
	dn     string
	serial string
}

type putRevokeSCEPCRTResponse struct {
	Err error
}

func (r putRevokeSCEPCRTResponse) error() error { return r.Err }
