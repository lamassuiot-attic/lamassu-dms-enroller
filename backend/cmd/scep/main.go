package main

import (
	"enroller/pkg/scep/api"
	"enroller/pkg/scep/auth"
	"enroller/pkg/scep/configs"
	"enroller/pkg/scep/discovery/consul"
	"enroller/pkg/scep/models/db"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	err, cfg := configs.NewConfig("scep")
	if err != nil {
		logger.Log("err", err, "msg", "Could not read environment configuration values")
		os.Exit(1)
	}

	connStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	db, err := db.NewDB("postgres", connStr, logger)
	if err != nil {
		logger.Log("err", err, "msg", "Could not start connection with CSRs database")
		os.Exit(1)
	}

	auth := auth.NewAuth(cfg.KeycloakHostname, cfg.KeycloakPort, cfg.KeycloakProtocol, cfg.KeycloakRealm, cfg.KeycloakCA)

	fieldKeys := []string{"method"}

	var s api.Service
	{
		s = api.NewSCEPService(db)
		s = api.LoggingMiddleware(logger)(s)
		s = api.NewInstrumentingMiddleware(
			kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "enroller",
				Subsystem: "scep_service",
				Name:      "request_count",
				Help:      "Number of requests received.",
			}, fieldKeys),
			kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "enroller",
				Subsystem: "scep_service",
				Name:      "request_latency_microseconds",
				Help:      "Total duration of requests in microseconds.",
			}, fieldKeys),
		)(s)
	}

	consulsd, err := consul.NewServiceDiscovery(cfg.ConsulProtocol, cfg.ConsulHost, cfg.ConsulPort, logger)
	if err != nil {
		logger.Log("err", err, "msg", "Could not start connection with Consul Service Discovery")
		os.Exit(1)
	}

	mux := http.NewServeMux()

	mux.Handle("/v1/", api.MakeHTTPHandler(s, log.With(logger, "component", "HTTPS"), auth))
	http.Handle("/", accessControl(mux, cfg.EnrollerUIProtocol, cfg.EnrollerUIHost, cfg.EnrollerUIPort))
	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTPS", "addr", "httpAddr")
		consulsd.Register("https", "scep", cfg.Port)
		errs <- http.ListenAndServeTLS(":"+cfg.Port, cfg.CertFile, cfg.KeyFile, nil)
	}()

	logger.Log("exit", <-errs)
	consulsd.Deregister()

}

func accessControl(h http.Handler, enrollerUIProtocol string, enrollerUIHost string, enrollerUIPort string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", enrollerUIProtocol+"://"+enrollerUIHost+":"+enrollerUIPort)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
