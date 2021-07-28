package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/lamassuiot/enroller/pkg/enroller/auth"
	"github.com/lamassuiot/enroller/pkg/enroller/models/csr"

	"github.com/gorilla/mux"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"

	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/nvellon/hal"
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

	r.Methods("POST").Path("/v1/csrs/{name}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostCSREndpoint),
		decodePostCSRRequest,
		encodePostCSRResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostCSR", logger)))...,
	))
	r.Methods("POST").Path("/v1/csrs/{name}/form").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostCSRFormEndpoint),
		decodePostCSRFormRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostCSRForm", logger)))...,
	))

	r.Methods("GET").Path("/v1/csrs").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRsEndpoint),
		decodeGetPendingCSRsRequest,
		encodeGetPendingCSRsResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetPendingCSRs", logger)))...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRDBEndpoint),
		decodeGetPendingCSRRequest,
		encodeGetPendingCSRResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetPendingCSRDB", logger)))...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}/file").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRFileEndpoint),
		decodeGetPendingCSRRequest,
		encodeGetPendingCSRFileResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetPendingCSRFile", logger)))...,
	))

	r.Methods("PUT").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PutChangeCSRStatusEndpoint),
		decodePutChangeCSRStatusRequest,
		encodePutChangeCSRStatusResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PutChangeCSRStatus", logger)))...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}/crt").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetCRTEndpoint),
		decodeGetCRTRequest,
		encodeGetCRTResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetCRT", logger)))...,
	))

	r.Methods("DELETE").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.DeleteCSREndpoint),
		decodeDeleteCSRRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "DeleteCSR", logger)))...,
	))

	return r
}

func decodeHealthRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req healthRequest
	return req, nil
}

func decodePostCSRFormRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, ErrEmptyDMSName
	}
	var csrForm csr.CSRForm
	json.NewDecoder(r.Body).Decode((&csrForm))
	if err != nil {
		return nil, errors.New("Cannot decode JSON request")
	}
	csrForm.Name = name
	req := postCSRFormRequest{
		CSRForm: csrForm,
	}
	return req, nil
}

func decodePostCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)

	var csrRequest postDirectCsr
	json.NewDecoder(r.Body).Decode((&csrRequest))

	name, ok := vars["name"]
	if !ok {
		return nil, ErrEmptyDMSName
	}
	req := postCSRRequest{
		data:    []byte(csrRequest.CSR),
		dmsName: name,
		url:     csrRequest.URL,
	}
	return req, nil
}

func decodeGetPendingCSRsRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req getPendingCSRsRequest
	return req, nil
}

func decodeGetPendingCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrInvalidID
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrInvalidIDFormat
	}
	return getPendingCSRRequest{ID: idNum}, nil
}

func decodePutChangeCSRStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrInvalidID
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrInvalidIDFormat
	}
	var c csr.CSR
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return nil, err
	}
	if c.Status == "" {
		return nil, ErrInvalidCSR
	}
	return putChangeCSRStatusRequest{CSR: c, ID: idNum}, nil

}

func decodeDeleteCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrInvalidID
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrInvalidIDFormat
	}
	return deleteCSRRequest{ID: idNum}, nil
}

func decodeGetCRTRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrInvalidID
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrInvalidIDFormat
	}
	return getCRTRequest{ID: idNum}, nil
}

func encodePostCSRResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(postCSRResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	url := "http://" + os.Getenv("ENROLLER_HOST") + os.Getenv("ENROLLER_PORT") + "/v1/csrs"
	csrHal := hal.NewResource(resp.CSR, url+strconv.Itoa(resp.CSR.Id))
	return json.NewEncoder(w).Encode(csrHal)
}

func encodeGetPendingCSRsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(getPendingCSRsResponse)
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	url := "http://" + os.Getenv("ENROLLER_HOST") + os.Getenv("ENROLLER_PORT") + "/v1/csrs"
	embedHal := hal.NewResource(resp.CSRs, url)
	for _, csr := range resp.CSRs.CSRs {
		csrHal := hal.NewResource(csr, url+strconv.Itoa(csr.Id))
		embedHal.Embed("csr", csrHal)
	}
	return json.NewEncoder(w).Encode(embedHal)
}

func encodeGetPendingCSRResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(getPendingCSRDBResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	url := "http://" + os.Getenv("ENROLLER_HOST") + os.Getenv("ENROLLER_PORT") + "/v1/csrs"
	csrHal := hal.NewResource(resp.CSR, url+strconv.Itoa(resp.CSR.Id))
	csrLink := hal.NewLink(url+strconv.Itoa(resp.CSR.Id)+"/file", hal.LinkAttr{
		"type": string("application/pkcs10"),
	})
	csrHal.AddLink("file", csrLink)
	return json.NewEncoder(w).Encode(csrHal)
}

func encodePutChangeCSRStatusResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(putChangeCSRsResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	url := "http://" + os.Getenv("ENROLLER_HOST") + os.Getenv("ENROLLER_PORT") + "/v1/csrs"
	csrHal := hal.NewResource(resp.CSR, url+strconv.Itoa(resp.CSR.Id))
	return json.NewEncoder(w).Encode(csrHal)
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

func encodeGetPendingCSRFileResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(getPendingCSRFileResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/pkcs10; charset=utf-8")
	w.Write(resp.Data)
	return nil
}

func encodeGetCRTResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(getCRTResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/x-pem-file; charset=utf-8")
	w.Write(resp.Data)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	http.Error(w, err.Error(), codeFrom(err))
}

func codeFrom(err error) int {
	switch err {
	case ErrInvalidCSR, ErrInvalidIDFormat, ErrInvalidApprobeOp, ErrInvalidDenyOp, ErrInvalidRevokeOp, ErrInvalidDeleteOp, ErrInvalidOperation:
		return http.StatusBadRequest
	case ErrInvalidID:
		return http.StatusNotFound
	case ErrIncorrectType:
		return http.StatusUnsupportedMediaType
	case jwt.ErrTokenExpired, jwt.ErrTokenInvalid, jwt.ErrTokenMalformed, jwt.ErrTokenNotActive, jwt.ErrTokenContextMissing, jwt.ErrUnexpectedSigningMethod:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
