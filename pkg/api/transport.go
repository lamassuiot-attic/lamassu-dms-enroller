package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/lamassuiot/dms-enroller/pkg/models/dms"

	"github.com/gorilla/mux"

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

func HTTPToContext(logger log.Logger) httptransport.RequestFunc {
	return func(ctx context.Context, req *http.Request) context.Context {
		// Try to join to a trace propagated in `req`.
		uberTraceId := req.Header.Values("Uber-Trace-Id")
		if uberTraceId != nil {
			logger = log.With(logger, "span_id", uberTraceId)
		} else {
			span := stdopentracing.SpanFromContext(ctx)
			logger = log.With(logger, "span_id", span)
		}
		return context.WithValue(ctx, "LamassuLogger", logger)
	}
}
func MakeHTTPHandler(s Service, logger log.Logger, otTracer stdopentracing.Tracer) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s, otTracer)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
		// httptransport.ServerBefore(jwt.HTTPToContext()),
	}

	r.Methods("GET").Path("/v1/health").Handler(httptransport.NewServer(
		e.HealthEndpoint,
		decodeHealthRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	r.Methods("POST").Path("/v1/{name}").Handler(httptransport.NewServer(
		e.PostCreateDMSEndpoint,
		decodePostCSRRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostCSR", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))
	r.Methods("POST").Path("/v1/{name}/form").Handler(httptransport.NewServer(
		e.PostCreateDMSFormEndpoint,
		decodePostCreateDMSFormRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PostCSRForm", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	r.Methods("GET").Path("/v1/").Handler(httptransport.NewServer(
		e.GetDMSsEndpoint,
		decodeGetDMSsRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetPendingCSRs", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	r.Methods("PUT").Path("/v1/{id}").Handler(httptransport.NewServer(
		e.PutChangeDMSStatusEndpoint,
		decodeputChangeDmsStatusRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "PutChangeCSRStatus", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	r.Methods("GET").Path("/v1/{id}/crt").Handler(httptransport.NewServer(
		e.GetCertificateEndpoint,
		decodeGetCRTRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetCRT", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	r.Methods("DELETE").Path("/v1/{id}").Handler(httptransport.NewServer(
		e.DeleteDMSEndpoint,
		decodeDeleteCSRRequest,
		encodeResponse,
		append(
			options,
			httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "DeleteCSR", logger)),
			httptransport.ServerBefore(HTTPToContext(logger)),
		)...,
	))

	return r
}

func decodeHealthRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req healthRequest
	return req, nil
}

func decodeGetDMSsRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req getDmsRequest
	return req, nil
}

func decodePostCreateDMSFormRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, ErrEmptyDMSName
	}
	var dmsForm dms.DmsCreationForm
	json.NewDecoder(r.Body).Decode((&dmsForm))
	if err != nil {
		return nil, errors.New("Cannot decode JSON request")
	}
	dmsForm.Name = name
	req := postDmsCreationFormRequest{
		DmsCreationForm: dmsForm,
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
		data:    csrRequest.CsrBase64Encoded,
		dmsName: name,
	}
	return req, nil
}

func decodeputChangeDmsStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrInvalidID
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, ErrInvalidIDFormat
	}
	var d dms.DMS
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return nil, err
	}
	if d.Status == "" {
		return nil, ErrInvalidCSR
	}
	return putChangeDmsStatusRequest{Dms: d, ID: idNum}, nil

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
