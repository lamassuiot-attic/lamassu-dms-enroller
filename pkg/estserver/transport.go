package estserver

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/lamassuiot/dms-enroller/pkg/estserver/mtls"
	lamassuest "github.com/lamassuiot/lamassu-est/pkg/server/api"

	"github.com/gorilla/mux"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type errorer interface {
	error() error
}

func MakeHTTPHandler(service lamassuest.Service, logger log.Logger, otTracer stdopentracing.Tracer) http.Handler {
	router := mux.NewRouter()
	endpoints := lamassuest.MakeServerEndpoints(service, otTracer)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(lamassuest.EncodeError),
		httptransport.ServerBefore(mtls.HTTPToContext()),
	}

	// MUST as per rfc7030
	router.Methods("GET").Path("/.well-known/est/cacerts").Handler(httptransport.NewServer(
		// mtls.NewParser()(endpoints.GetCAsEndpoint),
		endpoints.GetCAsEndpoint,
		lamassuest.DecodeRequest,
		lamassuest.EncodeGetCaCertsResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	router.Methods("POST").Path("/.well-known/est/{aps}/simpleenroll").Handler(httptransport.NewServer(
		mtls.NewParser()(endpoints.EnrollerEndpoint),
		//endpoints.EnrollerEndpoint,
		lamassuest.DecodeEnrollRequest,
		lamassuest.EncodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	router.Methods("POST").Path("/.well-known/est/simplereenroll").Handler(httptransport.NewServer(
		mtls.NewParser()(endpoints.ReenrollerEndpoint),
		//endpoints.ReenrollerEndpoint,
		lamassuest.DecodeReenrollRequest,
		lamassuest.EncodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	return router
}
