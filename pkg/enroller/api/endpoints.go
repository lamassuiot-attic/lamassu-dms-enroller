package api

import (
	"context"

	"github.com/lamassuiot/enroller/pkg/enroller/models/csr"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint             endpoint.Endpoint
	PostCSREndpoint            endpoint.Endpoint
	PostCSRFormEndpoint        endpoint.Endpoint
	GetPendingCSRsEndpoint     endpoint.Endpoint
	GetPendingCSRDBEndpoint    endpoint.Endpoint
	GetPendingCSRFileEndpoint  endpoint.Endpoint
	PutChangeCSRStatusEndpoint endpoint.Endpoint
	DeleteCSREndpoint          endpoint.Endpoint
	GetCRTEndpoint             endpoint.Endpoint
}

func MakeServerEndpoints(s Service, otTracer stdopentracing.Tracer) Endpoints {
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
		healthEndpoint = opentracing.TraceServer(otTracer, "Health")(healthEndpoint)
	}
	var postCSREndpoint endpoint.Endpoint
	{
		postCSREndpoint = MakePostCSREndpoint(s)
		postCSREndpoint = opentracing.TraceServer(otTracer, "PostCSR")(postCSREndpoint)
	}
	var postCSRFormEndpoint endpoint.Endpoint
	{
		postCSRFormEndpoint = MakePostCSRFormEndpoint(s)
		postCSRFormEndpoint = opentracing.TraceServer(otTracer, "PostCSRForm")(postCSRFormEndpoint)
	}
	var getPendingCSRsEndpoint endpoint.Endpoint
	{
		getPendingCSRsEndpoint = MakeGetPendingCSRsEndpoint(s)
		getPendingCSRsEndpoint = opentracing.TraceServer(otTracer, "GetPendingCSRs")(getPendingCSRsEndpoint)
	}
	var getPendingCSRDBEndpoint endpoint.Endpoint
	{
		getPendingCSRDBEndpoint = MakeGetPendingCSRDBEndpoint(s)
		getPendingCSRDBEndpoint = opentracing.TraceServer(otTracer, "GetPendingCSRDB")(getPendingCSRDBEndpoint)
	}
	var getPendingCSRFileEndpoint endpoint.Endpoint
	{
		getPendingCSRFileEndpoint = MakeGetPendingCSRFileEndpoint(s)
		getPendingCSRFileEndpoint = opentracing.TraceServer(otTracer, "GetPendingCSRFile")(getPendingCSRFileEndpoint)
	}
	var putChangeCSRStatusEndpoint endpoint.Endpoint
	{
		putChangeCSRStatusEndpoint = MakePutChangeCSRStatusEndpoint(s)
		putChangeCSRStatusEndpoint = opentracing.TraceServer(otTracer, "PutChangeCSRStatus")(putChangeCSRStatusEndpoint)
	}
	var deleteCSREndpoint endpoint.Endpoint
	{
		deleteCSREndpoint = MakeDeleteCSREndpoint(s)
		deleteCSREndpoint = opentracing.TraceServer(otTracer, "DeleteCSR")(deleteCSREndpoint)
	}
	var getCRTEndpoint endpoint.Endpoint
	{
		getCRTEndpoint = MakeGetCTREndpoint(s)
		getCRTEndpoint = opentracing.TraceServer(otTracer, "GetCRT")(getCRTEndpoint)
	}

	return Endpoints{
		HealthEndpoint:             healthEndpoint,
		PostCSREndpoint:            postCSREndpoint,
		PostCSRFormEndpoint:        postCSRFormEndpoint,
		GetPendingCSRsEndpoint:     getPendingCSRsEndpoint,
		GetPendingCSRDBEndpoint:    getPendingCSRDBEndpoint,
		GetPendingCSRFileEndpoint:  getPendingCSRFileEndpoint,
		PutChangeCSRStatusEndpoint: putChangeCSRStatusEndpoint,
		DeleteCSREndpoint:          deleteCSREndpoint,
		GetCRTEndpoint:             getCRTEndpoint,
	}
}

func MakeHealthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return healthResponse{Healthy: healthy}, nil
	}
}

func MakePostCSREndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postCSRRequest)
		csr, e := s.PostCSR(ctx, req.data, req.dmsName, req.url)
		return postCSRResponse{CSR: csr, Err: e}, nil
	}
}

func MakePostCSRFormEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postCSRFormRequest)
		privKey, csr, e := s.PostCSRForm(ctx, req.CSRForm)
		return postCSRFormResponse{PrivKey: privKey, CSR: csr, Err: e}, nil
	}
}

func MakeGetPendingCSRsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(getPendingCSRsRequest)
		csrs := s.GetPendingCSRs(ctx)
		return getPendingCSRsResponse{CSRs: csrs}, nil
	}
}

func MakeGetPendingCSRDBEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getPendingCSRRequest)
		csr, err := s.GetPendingCSRDB(ctx, req.ID)
		return getPendingCSRDBResponse{CSR: csr, Err: err}, nil
	}
}

func MakeGetPendingCSRFileEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getPendingCSRRequest)
		data, err := s.GetPendingCSRFile(ctx, req.ID)
		return getPendingCSRFileResponse{Data: data, Err: err}, nil
	}
}

func MakePutChangeCSRStatusEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putChangeCSRStatusRequest)
		csr, err := s.PutChangeCSRStatus(ctx, req.CSR, req.ID)
		return putChangeCSRsResponse{CSR: csr, Err: err}, nil
	}
}

func MakeDeleteCSREndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteCSRRequest)
		err = s.DeleteCSR(ctx, req.ID)
		return deleteCSRResponse{Err: err}, nil
	}
}

func MakeGetCTREndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getCRTRequest)
		data, err := s.GetCRT(ctx, req.ID)
		return getCRTResponse{Data: data, Err: err}, nil
	}
}

type healthRequest struct{}

type healthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"err,omitempty"`
}

type getCRTRequest struct {
	ID int
}

type getCRTResponse struct {
	Data []byte
	Err  error
}

type postCSRRequest struct {
	data    []byte
	dmsName string
	url     string
}
type postCSRFormRequest struct {
	CSRForm csr.CSRForm
}

type postCSRResponse struct {
	CSR csr.CSR `json:"csr,omitempty"`
	Err error   `json:"err,omitempty"`
}
type postCSRFormResponse struct {
	CSR     csr.CSR `json:"csr,omitempty"`
	PrivKey string  `json:"priv_key,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r postCSRResponse) error() error { return r.Err }

type getPendingCSRsRequest struct{}

type getPendingCSRsResponse struct {
	CSRs csr.CSRs `json:"CSRs,omitempty"`
}

type getPendingCSRRequest struct {
	ID int
}

type getPendingCSRDBResponse struct {
	CSR csr.CSR `json:"CSR,omitempty"`
	Err error   `json:"err,omitempty"`
}

func (r getPendingCSRDBResponse) error() error { return r.Err }

type getPendingCSRFileResponse struct {
	Data []byte
	Err  error
}

type postDirectCsr struct {
	CSR string `json:"csr"`
	URL string `json:"url"`
}

type putChangeCSRStatusRequest struct {
	CSR csr.CSR
	ID  int
}

type putChangeCSRsResponse struct {
	CSR csr.CSR
	Err error
}

func (r putChangeCSRsResponse) error() error { return r.Err }

type deleteCSRRequest struct {
	ID int
}

type deleteCSRResponse struct {
	Err error
}

func (r deleteCSRResponse) error() error { return r.Err }
