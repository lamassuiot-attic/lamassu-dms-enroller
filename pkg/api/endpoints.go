package api

import (
	"context"

	"github.com/lamassuiot/dms-enroller/pkg/models/dms"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint             endpoint.Endpoint
	PostCreateDMSEndpoint      endpoint.Endpoint
	PostCreateDMSFormEndpoint  endpoint.Endpoint
	PutChangeDMSStatusEndpoint endpoint.Endpoint
	DeleteDMSEndpoint          endpoint.Endpoint
	GetDMSsEndpoint            endpoint.Endpoint
	GetCertificateEndpoint     endpoint.Endpoint
}

func MakeServerEndpoints(s Service, otTracer stdopentracing.Tracer) Endpoints {
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
		healthEndpoint = opentracing.TraceServer(otTracer, "Health")(healthEndpoint)
	}
	var postCreateDMSEndpoint endpoint.Endpoint
	{
		postCreateDMSEndpoint = MakeCreateDMSEndpoint(s)
		postCreateDMSEndpoint = opentracing.TraceServer(otTracer, "CreateDMS")(postCreateDMSEndpoint)
	}
	var postCreateDMSFormEndpoint endpoint.Endpoint
	{
		postCreateDMSFormEndpoint = MakeCreateDMSFormEndpoint(s)
		postCreateDMSFormEndpoint = opentracing.TraceServer(otTracer, "CreateDMSForm")(postCreateDMSFormEndpoint)
	}

	var getDMSsEndpoint endpoint.Endpoint
	{
		getDMSsEndpoint = MakeGetDMSsEndpoint(s)
		getDMSsEndpoint = opentracing.TraceServer(otTracer, "GetDMSs")(getDMSsEndpoint)
	}
	var putChangeDMSStatusEndpoint endpoint.Endpoint
	{
		putChangeDMSStatusEndpoint = MakeChangeDMSStatusEndpoint(s)
		putChangeDMSStatusEndpoint = opentracing.TraceServer(otTracer, "ChangeDMSStatus")(putChangeDMSStatusEndpoint)
	}
	var deleteDmsEndpoint endpoint.Endpoint
	{
		deleteDmsEndpoint = MakeDeleteDMSEndpoint(s)
		deleteDmsEndpoint = opentracing.TraceServer(otTracer, "DeleteDMS")(deleteDmsEndpoint)
	}
	var getCertificateEndpoint endpoint.Endpoint
	{
		getCertificateEndpoint = MakeGetCertificateEndpoint(s)
		getCertificateEndpoint = opentracing.TraceServer(otTracer, "GetCertificate")(getCertificateEndpoint)
	}

	return Endpoints{
		HealthEndpoint:             healthEndpoint,
		PostCreateDMSEndpoint:      postCreateDMSEndpoint,
		PostCreateDMSFormEndpoint:  postCreateDMSFormEndpoint,
		PutChangeDMSStatusEndpoint: putChangeDMSStatusEndpoint,
		DeleteDMSEndpoint:          deleteDmsEndpoint,
		GetDMSsEndpoint:            getDMSsEndpoint,
		GetCertificateEndpoint:     getCertificateEndpoint,
	}
}

func MakeHealthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return healthResponse{Healthy: healthy}, nil
	}
}

func MakeCreateDMSEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postCSRRequest)
		dms, e := s.CreateDMS(ctx, req.data, req.dmsName)
		return dms, e
	}
}

func MakeCreateDMSFormEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postDmsCreationFormRequest)
		privKey, dms, e := s.CreateDMSForm(ctx, req.DmsCreationForm)
		return postDmsCreationFormResponse{PrivKey: privKey, Dms: dms, Err: e}, nil
	}
}

func MakeGetDMSsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(getDmsRequest)
		dmss, err := s.GetDMSs(ctx)
		return dmss, err
	}
}

func MakeChangeDMSStatusEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(putChangeDmsStatusRequest)
		dms, err := s.UpdateDMSStatus(ctx, req.Dms, req.ID)
		return dms, err
	}
}

func MakeDeleteDMSEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteCSRRequest)
		err = s.DeleteDMS(ctx, req.ID)
		return deleteCSRResponse{Err: err}, nil
	}
}

func MakeGetCertificateEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getCRTRequest)
		data, err := s.GetDMSCertificate(ctx, req.ID)
		return getCRTResponse{Data: data.Raw, Err: err}, nil
	}
}

type healthRequest struct{}

type healthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"err,omitempty"`
}
type getDmsRequest struct{}

type getCRTRequest struct {
	ID int
}

type getCRTResponse struct {
	Data []byte
	Err  error
}

type postCSRRequest struct {
	data    string
	dmsName string
	url     string
}
type postDmsCreationFormRequest struct {
	DmsCreationForm dms.DmsCreationForm
}

type postDmsResponse struct {
	Dms dms.DMS `json:"dms,omitempty"`
	Err error   `json:"err,omitempty"`
}
type postDmsCreationFormResponse struct {
	Dms     dms.DMS `json:"dms,omitempty"`
	PrivKey string  `json:"priv_key,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r postDmsResponse) error() error { return r.Err }

type getPendingCSRFileResponse struct {
	Data []byte
	Err  error
}

type postDirectCsr struct {
	CsrBase64Encoded string `json:"csr"`
}

type putChangeDmsStatusRequest struct {
	Dms dms.DMS
	ID  int
}

type putChangeCSRsResponse struct {
	Dms dms.DMS
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
