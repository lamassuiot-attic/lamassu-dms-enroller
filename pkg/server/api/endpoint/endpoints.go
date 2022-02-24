package endpoint

import (
	"context"
	"crypto/x509"
	"math"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-playground/validator/v10"
	dmsenrrors "github.com/lamassuiot/dms-enroller/pkg/server/api/errors"
	"github.com/lamassuiot/dms-enroller/pkg/server/api/service"
	"github.com/lamassuiot/dms-enroller/pkg/server/models/dms"
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

func MakeServerEndpoints(s service.Service, otTracer stdopentracing.Tracer) Endpoints {
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

func MakeHealthEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return HealthResponse{Healthy: healthy}, nil
	}
}

func MakeCreateDMSEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(PostCSRRequest)
		err = ValidatetPostCSRRequest(req)
		if err != nil {
			valError := dmsenrrors.ValidationError{
				Msg: err.Error(),
			}
			return nil, &valError
		}
		dms, e := s.CreateDMS(ctx, req.Csr, req.DmsName)
		return dms, e
	}
}

func MakeCreateDMSFormEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(PostDmsCreationFormRequest)
		err = ValidatePostDmsCreationFormRequest(req)
		if err != nil {
			valError := dmsenrrors.ValidationError{
				Msg: err.Error(),
			}
			return nil, &valError
		}
		privKey, dms, e := s.CreateDMSForm(ctx, dms.Subject(req.Subject), dms.PrivateKeyMetadata(req.KeyMetadata), req.Url, req.DmsName)
		return PostDmsCreationFormResponse{PrivKey: privKey, Dms: dms, Err: e}, nil
	}
}

func MakeGetDMSsEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_ = request.(GetDmsRequest)
		dmss, err := s.GetDMSs(ctx)
		return dmss, err
	}
}

func MakeChangeDMSStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(PutChangeDmsStatusRequest)
		err = ValidatetPutChangeDmsStatusRequest(req)
		if err != nil {
			valError := dmsenrrors.ValidationError{
				Msg: err.Error(),
			}
			return nil, &valError
		}
		dms, err := s.UpdateDMSStatus(ctx, req.Status, req.ID)
		return dms, err
	}
}

func MakeDeleteDMSEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(DeleteCSRRequest)
		err = s.DeleteDMS(ctx, req.ID)
		if err != nil {
			return "", err
		} else {
			return "OK", err
		}
	}
}

func MakeGetCertificateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(GetCRTRequest)
		data, err := s.GetDMSCertificate(ctx, req.ID)
		return GetCRTResponse{Data: data, Err: err}, nil
	}
}

type HealthRequest struct{}

type HealthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"err,omitempty"`
}
type GetDmsRequest struct{}

type GetCRTRequest struct {
	ID int
}

type GetCRTResponse struct {
	Data *x509.Certificate
	Err  error
}

type PostCSRRequest struct {
	Csr     string `json:"csr" validate:"required"`
	DmsName string `json:"dms_name" validate:"required"`
	Url     string `json:"url" validate:"required"`
}

func ValidatetPostCSRRequest(request PostCSRRequest) error {
	validate := validator.New()
	return validate.Struct(request)
}

type PostDmsCreationFormRequest struct {
	DmsName string `json:"dms_name" validate:"required"`
	Subject struct {
		CN string `json:"common_name"`
		O  string `json:"organization"`
		OU string `json:"organization_unit"`
		C  string `json:"country"`
		ST string `json:"state"`
		L  string `json:"locality"`
	} `json:"subject"`
	KeyMetadata struct {
		KeyType string `json:"type" validate:"oneof='rsa' 'ecdsa'"`
		KeyBits int    `json:"bits" validate:"required"`
	} `json:"key_metadata" validate:"required"`
	Url string `json:"url" validate:"required"`
}

func ValidatePostDmsCreationFormRequest(request PostDmsCreationFormRequest) error {
	CreateCARequestStructLevelValidation := func(sl validator.StructLevel) {
		req := sl.Current().Interface().(PostDmsCreationFormRequest)
		switch req.KeyMetadata.KeyType {
		case "rsa":
			if math.Mod(float64(req.KeyMetadata.KeyBits), 1024) != 0 && req.KeyMetadata.KeyBits < 2048 {
				sl.ReportError(req.KeyMetadata.KeyBits, "bits", "Bits", "bits1024multipleAndGt2048", "")
			}
		case "ecdsa":
			if req.KeyMetadata.KeyBits < 160 || req.KeyMetadata.KeyBits > 512 {
				sl.ReportError(req.KeyMetadata.KeyBits, "bits", "Bits", "bitsEcdsaMultiple", "")
			}
		}
	}

	validate := validator.New()
	validate.RegisterStructValidation(CreateCARequestStructLevelValidation, PostDmsCreationFormRequest{})
	return validate.Struct(request)
}

type PostDmsResponse struct {
	Dms dms.DMS `json:"dms,omitempty"`
	Err error   `json:"err,omitempty"`
}
type PostDmsCreationFormResponse struct {
	Dms     dms.DMS `json:"dms,omitempty"`
	PrivKey string  `json:"priv_key,omitempty"`
	Err     error   `json:"err,omitempty"`
}

func (r PostDmsResponse) error() error { return r.Err }

type GetPendingCSRFileResponse struct {
	Data []byte
	Err  error
}

type PostDirectCsr struct {
	CsrBase64Encoded string `json:"csr"`
}

type PutChangeDmsStatusRequest struct {
	Status string `json:"status" validate:"required"`
	ID     int    `validate:"required"`
}

func ValidatetPutChangeDmsStatusRequest(request PutChangeDmsStatusRequest) error {
	validate := validator.New()
	return validate.Struct(request)
}

type PutChangeCSRsResponse struct {
	Dms dms.DMS
	Err error
}

func (r PutChangeCSRsResponse) error() error { return r.Err }

type DeleteCSRRequest struct {
	ID int
}

type DeleteCSRResponse struct {
	Err error
}

func (r DeleteCSRResponse) error() error { return r.Err }
