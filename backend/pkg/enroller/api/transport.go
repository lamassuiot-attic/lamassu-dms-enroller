package api

import (
	"context"
	"encoding/json"
	"enroller/pkg/enroller/auth"
	"enroller/pkg/enroller/crypto"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	stdjwt "github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/auth/jwt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/nvellon/hal"
)

type errorer interface {
	error() error
}

var claims = &auth.KeycloakClaims{}

func MakeHTTPHandler(s Service, logger log.Logger, auth auth.Auth) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(jwt.HTTPToContext()),
	}

	r.Methods("POST").Path("/v1/csrs").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PostCSREndpoint),
		decodePostCSRRequest,
		encodePostCSRResponse,
		options...,
	))

	r.Methods("GET").Path("/v1/csrs").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRsEndpoint),
		decodeGetPendingCSRsRequest,
		encodeGetPendingCSRsResponse,
		options...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRDBEndpoint),
		decodeGetPendingCSRRequest,
		encodeGetPendingCSRResponse,
		options...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}/file").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetPendingCSRFileEndpoint),
		decodeGetPendingCSRRequest,
		encodeGetPendingCSRFileResponse,
		options...,
	))

	r.Methods("PUT").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.PutChangeCSRStatusEndpoint),
		decodePutChangeCSRStatusRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/v1/csrs/{id}/crt").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.GetCRTEndpoint),
		decodeGetCRTRequest,
		encodeGetCRTResponse,
		options...,
	))

	r.Methods("DELETE").Path("/v1/csrs/{id}").Handler(httptransport.NewServer(
		jwt.NewParser(auth.Kf, stdjwt.SigningMethodRS256, auth.KeycloakClaimsFactory)(e.DeleteCSREndpoint),
		decodeDeleteCSRRequest,
		encodeResponse,
		options...,
	))

	return r
}

func decodePostCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/pkcs10" {
		return nil, ErrIncorrectType
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	req := postCSRRequest{
		data: data,
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
		return nil, ErrBadRouting
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrBadRouting
	}
	return getPendingCSRRequest{ID: idNum}, nil
}

func decodePutChangeCSRStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrBadRouting
	}
	var csr crypto.CSR
	if err := json.NewDecoder(r.Body).Decode(&csr); err != nil {
		return nil, err
	}
	if csr.Status == "" {
		return nil, ErrIncorrectContent
	}
	return putChangeCSRStatusRequest{CSR: csr, ID: idNum}, nil

}

func decodeDeleteCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrBadRouting
	}
	return deleteCSRRequest{ID: idNum}, nil
}

func decodeGetCRTRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrBadRouting
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrBadRouting
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
	csrHal := hal.NewResource(resp.CSR, "http://localhost:8080/v1/csrs")
	csrHal.AddNewLink("csr", "http://localhost:8080/v1/csrs/"+strconv.Itoa(resp.CSR.Id))
	return json.NewEncoder(w).Encode(csrHal)
}

func encodeGetPendingCSRsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(getPendingCSRsResponse)
	w.Header().Set("Content-Type", "application/hal+json; charset=utf-8")
	embedHal := hal.NewResource(resp.CSRs, "http://localhost:8080/v1/csrs")
	for _, csr := range resp.CSRs.CSRs {
		csrHal := hal.NewResource(csr, "http://localhost:8080/v1/csrs/"+strconv.Itoa(csr.Id))
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
	csrHal := hal.NewResource(resp.CSR, "http://localhost:8080/v1/csrs/"+strconv.Itoa(resp.CSR.Id))
	csrLink := hal.NewLink("http://localhost:8080/v1/csrs/"+strconv.Itoa(resp.CSR.Id)+"/file", hal.LinkAttr{
		"type": string("application/pkcs10"),
	})
	csrHal.AddLink("file", csrLink)
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
	case ErrIncorrectContent, ErrBadRouting:
		return http.StatusBadRequest
	case ErrIncorrectType:
		return http.StatusUnsupportedMediaType
	case jwt.ErrTokenExpired, jwt.ErrTokenInvalid, jwt.ErrTokenMalformed, jwt.ErrTokenNotActive, jwt.ErrTokenContextMissing, jwt.ErrUnexpectedSigningMethod:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
