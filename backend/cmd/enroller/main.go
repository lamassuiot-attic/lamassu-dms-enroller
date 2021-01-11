package main

import (
	"enroller/pkg/enroller/api"
	"enroller/pkg/enroller/auth"
	"enroller/pkg/enroller/configs"
	"enroller/pkg/enroller/discovery/consul"
	certsdb "enroller/pkg/enroller/models/certs/store/db"
	certsfile "enroller/pkg/enroller/models/certs/store/file"
	csrdb "enroller/pkg/enroller/models/csr/store/db"
	csrfile "enroller/pkg/enroller/models/csr/store/file"
	secrets "enroller/pkg/enroller/secrets/file"
	"flag"
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
	err, cfg := configs.NewConfig("enroller")
	if err != nil {
		panic(err)
	}

	csrConnStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	csrdb, err := csrdb.NewDB("postgres", csrConnStr)
	if err != nil {
		panic(err)
	}
	csrfile := csrfile.NewFile(cfg.HomePath)

	certsConnStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	certsdb, err := certsdb.NewDB("postgres", certsConnStr)
	if err != nil {
		panic(err)
	}
	certsfile := certsfile.NewFile(cfg.HomePath)

	auth := auth.NewAuth(cfg.KeycloakHostname, cfg.KeycloakPort, cfg.KeycloakProtocol, cfg.KeycloakRealm, cfg.KeycloakCA)

	secrets := secrets.NewFile(cfg.CACertFile, cfg.CAKeyFile, cfg.OCSPServer, certsdb)

	var (
		httpAddr = flag.String("http.addr", ":"+cfg.Port, "HTTPS listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	fieldKeys := []string{"method"}

	var s api.Service
	{
		s = api.NewEnrollerService(csrdb, csrfile, certsdb, certsfile, secrets, cfg.HomePath)
		s = api.LoggingMiddleware(logger)(s)
		s = api.NewInstrumentingMiddleware(
			kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "enroller",
				Subsystem: "enroller_service",
				Name:      "request_count",
				Help:      "Number of requests received.",
			}, fieldKeys),
			kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "enroller",
				Subsystem: "enroller_service",
				Name:      "request_latency_microseconds",
				Help:      "Total duration of requests in microseconds.",
			}, fieldKeys),
		)(s)
	}

	consulsd, err := consul.NewServiceDiscovery(cfg.ConsulProtocol, cfg.ConsulHost, cfg.ConsulPort, logger)
	if err != nil {
		panic(err)
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
		logger.Log("transport", "HTTPS", "addr", "httpsAddr")
		consulsd.Register("https", "enroller", cfg.Port)
		errs <- http.ListenAndServeTLS(*httpAddr, cfg.CertFile, cfg.KeyFile, nil)
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
