package api

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/lamassuiot/enroller/pkg/devices/auth"
	"github.com/lamassuiot/enroller/pkg/devices/models/device"

	"github.com/gorilla/mux"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"

	stdopentracing "github.com/opentracing/opentracing-go"
)

type errorer interface {
	error() error
}

var claims = &auth.KeycloakClaims{}

func MakeHTTPHandler(s Service, logger log.Logger, auth auth.Auth, otTracer stdopentracing.Tracer) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s, otTracer)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(jwt.HTTPToContext()),
	}

	r.Methods("GET").Path("/v1/health").Handler(httptransport.NewServer(
		e.HealthEndpoint,
		decodeHealthRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	r.Methods("POST").Path("/v1/devices").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostDeviceEndpoint),
		decodePostDeviceRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostDevice", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDevices),
		decodeRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDevices", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices/{deviceId}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDeviceById),
		decodeGetDeviceById,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDeviceById", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices/dms/{dmsId}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDevicesByDMS),
		decodeGetDevicesByDMSRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDevicesByDMS", logger)))...,
	))

	r.Methods("DELETE").Path("/v1/devices/{deviceId}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.DeleteDevice),
		decodeDeleteDeviceRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "DeleteDevice", logger)))...,
	))

	r.Methods("POST").Path("/v1/devices/{deviceId}/issue").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostIssue),
		decodePostIssueRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostIssue", logger)))...,
	))

	r.Methods("POST").Path("/v1/devices/{deviceId}/issue/dms").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostIssuedViaDMS),
		decodePostIssueViaDMSRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostIssuedViaDMS", logger)))...,
	))

	r.Methods("POST").Path("/v1/devices/{deviceId}/issue/defaults").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostIssueUsingDefaults),
		decodePostIssueRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostIssue", logger)))...,
	))

	r.Methods("DELETE").Path("/v1/devices/{deviceId}/revoke").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.DeleteRevoke),
		decodedecodeDeleteRevokeRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "DeleteRevoke", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices/{deviceId}/logs").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDeviceLogs),
		decodedecodeGetDeviceLogsRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDeviceLogs", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices/{deviceId}/cert").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDeviceLogs),
		decodedecodeGetDeviceCertRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDeviceCert", logger)))...,
	))

	r.Methods("GET").Path("/v1/devices/{deviceId}/cert-history").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetDeviceLogs),
		decodedecodeGetDeviceCertHistoryRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetDeviceCertHistory", logger)))...,
	))

	return r
}

func decodeHealthRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req healthRequest
	return req, nil
}

func decodeRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req healthRequest
	return req, nil
}

func decodePostDeviceRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var device device.Device
	json.NewDecoder(r.Body).Decode((&device))
	if err != nil {
		return nil, errors.New("Cannot decode JSON request")
	}
	req := postDeviceRequest{
		Device: device,
	}
	return req, nil
}

func decodeGetDeviceById(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDMSId
	}
	return getDevicesByIdRequest{Id: id}, nil
}

func decodeGetDevicesByDMSRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["dmsId"]
	if !ok {
		return nil, ErrInvalidDMSId
	}
	return getDevicesByDMSRequest{Id: id}, nil
}

func decodeDeleteDeviceRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDMSId
	}
	return deleteDeviceRequest{Id: id}, nil
}
func decodePostIssueRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/pkcs10" {
		return nil, ErrIncorrectType
	}
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, ErrEmptyBody
	}
	req := postIssueRequest{
		Id:  id,
		Csr: data,
	}
	return req, nil
}
func decodePostIssueViaDMSRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var issueViaDmsRequest postIssueViaDMSRequest
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	json.NewDecoder(r.Body).Decode((&issueViaDmsRequest))
	if err != nil {
		return nil, errors.New("Cannot decode JSON request")
	}
	issueViaDmsRequest.DeviceId = id
	return issueViaDmsRequest, nil
}
func decodedecodeDeleteRevokeRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	return deleteRevokeRequest{Id: id}, nil
}
func decodedecodeGetDeviceLogsRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	return getDeviceLogsRequest{Id: id}, nil
}
func decodedecodeGetDeviceCertRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	return getDeviceCertRequest{Id: id}, nil
}
func decodedecodeGetDeviceCertHistoryRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["deviceId"]
	if !ok {
		return nil, ErrInvalidDeviceId
	}
	return getDeviceCertHistoryRequest{Id: id}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	http.Error(w, err.Error(), codeFrom(err))
}

func codeFrom(err error) int {
	switch err {
	case ErrInvalidDeviceRequest, ErrInvalidOperation:
		return http.StatusBadRequest
	case ErrIncorrectType:
		return http.StatusUnsupportedMediaType
	case jwt.ErrTokenExpired, jwt.ErrTokenInvalid, jwt.ErrTokenMalformed, jwt.ErrTokenNotActive, jwt.ErrTokenContextMissing, jwt.ErrUnexpectedSigningMethod:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
