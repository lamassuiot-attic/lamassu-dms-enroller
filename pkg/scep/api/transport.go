package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lamassuiot/enroller/pkg/scep/auth"
	"github.com/lamassuiot/enroller/pkg/scep/crypto"

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

	r.Methods("GET").Path("/v1/scep").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetSCEPCRTsEndpoint),
		decodeGetSCEPCRTsRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetSCEPCRT", logger)))...,
	))

	r.Methods("PUT").Path("/v1/scep/{serial}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PutRevokeSCEPCRTEndpoint),
		decodePutRevokeSCEPCRTRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "RevokeSCEPCRT", logger)))...,
	))

	return r

}

func decodeHealthRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req healthRequest
	return req, nil
}

func decodeGetSCEPCRTsRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req getSCEPCRTsRequest
	return req, nil
}

func decodePutRevokeSCEPCRTRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	serial, ok := vars["serial"]
	if !ok {
		return nil, ErrInvalidDNOrSerial
	}
	var crt crypto.CRT
	if err := json.NewDecoder(r.Body).Decode(&crt); err != nil {
		return nil, err
	}
	if crt.DN == "" {
		return nil, ErrInvalidCert
	}

	return putRevokeSCEPCRTRequest{dn: crt.DN, serial: serial}, nil

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
	case ErrInvalidCert, ErrInvalidRevokeOp:
		return http.StatusBadRequest
	case ErrInvalidDNOrSerial:
		return http.StatusNotFound
	case jwt.ErrTokenExpired, jwt.ErrTokenInvalid, jwt.ErrTokenMalformed, jwt.ErrTokenNotActive, jwt.ErrTokenContextMissing, jwt.ErrUnexpectedSigningMethod:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
