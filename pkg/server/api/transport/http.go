package transport

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lamassuiot/dms-enroller/pkg/server/api/endpoint"
	dmsErrors "github.com/lamassuiot/dms-enroller/pkg/server/api/errors"
	"github.com/lamassuiot/dms-enroller/pkg/server/api/service"
	"github.com/lamassuiot/dms-enroller/pkg/server/utils"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"

	stdopentracing "github.com/opentracing/opentracing-go"
)

type errorer interface {
	error() error
}

func ErrMissingDMSID() error {
	return &dmsErrors.GenericError{
		Message:    "DMS ID not specified",
		StatusCode: 400,
	}
}
func ErrMissingDMSName() error {
	return &dmsErrors.GenericError{
		Message:    "DMS name not specified",
		StatusCode: 400,
	}
}
func ErrMissingDMSStatus() error {
	return &dmsErrors.GenericError{
		Message:    "DMS status not specified",
		StatusCode: 400,
	}
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
		return context.WithValue(ctx, utils.LamassuLoggerContextKey, logger)
	}
}
func MakeHTTPHandler(s service.Service, logger log.Logger, otTracer stdopentracing.Tracer) http.Handler {
	r := mux.NewRouter()
	e := endpoint.MakeServerEndpoints(s, otTracer)
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
		encodeGetCRTResponse,
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
	var req endpoint.HealthRequest
	return req, nil
}

func decodeGetDMSsRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoint.GetDmsRequest
	return req, nil
}

func decodePostCreateDMSFormRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	name, ok := vars["name"]
	if !ok {
		return nil, ErrMissingDMSName()
	}
	var dmsForm endpoint.PostDmsCreationFormRequest
	json.NewDecoder(r.Body).Decode((&dmsForm))
	if err != nil {
		return nil, errors.New("Cannot decode JSON request")
	}
	dmsForm.DmsName = name
	return dmsForm, nil
}

func decodePostCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)

	var csrRequest endpoint.PostDirectCsr
	json.NewDecoder(r.Body).Decode((&csrRequest))

	name, ok := vars["name"]
	if !ok {
		return nil, ErrMissingDMSName()
	}

	req := endpoint.PostCSRRequest{
		Csr:     csrRequest.CsrBase64Encoded,
		DmsName: name,
	}
	return req, nil
}

func decodeputChangeDmsStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrMissingDMSID()
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	var statusRequest endpoint.PutChangeDmsStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&statusRequest); err != nil {
		return nil, err
	}
	if statusRequest.Status == "" {
		return nil, ErrMissingDMSStatus()
	}
	statusRequest.ID = idNum
	return statusRequest, nil

}

func decodeDeleteCSRRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrMissingDMSID()
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return endpoint.DeleteCSRRequest{ID: idNum}, nil
}

func decodeGetCRTRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, ErrMissingDMSID()
	}
	idNum, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	return endpoint.GetCRTRequest{ID: idNum}, nil
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
	resp := response.(endpoint.GetPendingCSRFileResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	w.Header().Set("Content-Type", "application/pkcs10; charset=utf-8")
	w.Write(resp.Data)
	return nil
}

func encodeGetCRTResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(endpoint.GetCRTResponse)
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}
	cert := resp.Data
	var cb []byte
	cb = append(cb, cert.Raw...)
	b := pem.Block{Type: "CERTIFICATE", Bytes: cb}
	body := pem.EncodeToMemory(&b)

	body = []byte(utils.EcodeB64(string(body)))

	w.Header().Set("Content-Type", "application/pkcs7-mime; smime-type=certs-only")
	w.Header().Set("Content-Transfer-Encoding", "base64")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	http.Error(w, err.Error(), codeFrom(err))
}

func codeFrom(err error) int {
	switch e := err.(type) {
	case *dmsErrors.ValidationError:
		return http.StatusBadRequest
	case *dmsErrors.DuplicateResourceError:
		return http.StatusConflict
	case *dmsErrors.ResourceNotFoundError:
		return http.StatusNotFound
	case *dmsErrors.GenericError:
		return e.StatusCode
	default:
		return http.StatusInternalServerError
	}
}
