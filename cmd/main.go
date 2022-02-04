package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-openapi/runtime/middleware"
	"github.com/lamassuiot/dms-enroller/pkg/api"
	"github.com/lamassuiot/dms-enroller/pkg/utils"

	"github.com/lamassuiot/dms-enroller/pkg/config"
	dmsdb "github.com/lamassuiot/dms-enroller/pkg/models/dms/store/db"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/client"

	"github.com/lamassuiot/dms-enroller/pkg/docs"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"gopkg.in/yaml.v2"
)

func main() {
	var logger log.Logger
	{
		logger = log.NewJSONLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, level.AllowInfo())
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	err, cfg := config.NewConfig("")
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not read environment configuration values")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Environment configuration values loaded")

	dmsConnStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	dmsDb, err := dmsdb.NewDB("postgres", dmsConnStr, logger)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not start connection with DMS Enroller database. Will sleep for 5 seconds and exit the program")
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "Connection established with DMS Enroller database")

	jcfg, err := jaegercfg.FromEnv()
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not load Jaeger configuration values fron environment")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Jaeger configuration values loaded")
	tracer, closer, err := jcfg.NewTracer()
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not start Jaeger tracer")
		os.Exit(1)
	}
	defer closer.Close()
	level.Info(logger).Log("msg", "Jaeger tracer started")

	fieldKeys := []string{"method", "error"}

	lamassuCaClient, err := lamassucaclient.NewLamassuCaClient(cfg.LamassuCAAddress, cfg.LamassuCACertFile, cfg.CertFile, cfg.KeyFile, logger)

	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create lamassu CA client")
		os.Exit(1)
	}

	var s api.Service
	{
		s = api.NewEnrollerService(dmsDb, &lamassuCaClient, logger)
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
	openapiSpec := docs.NewOpenAPI3(cfg)

	openapiSpecJsonData, _ := json.Marshal(&openapiSpec)
	openapiSpecYamlData, _ := yaml.Marshal(&openapiSpec)

	err = os.MkdirAll("docs", 0744)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create openapiv3 docs dir")
		os.Exit(1)
	}

	err = os.WriteFile(path.Join("docs", "openapiv3.json"), openapiSpecJsonData, 0644)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create openapiv3 JSON spec file")
		os.Exit(1)
	}

	err = os.WriteFile(path.Join("docs", "openapiv3.yaml"), openapiSpecYamlData, 0644)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create openapiv3 YAML spec file")
		os.Exit(1)
	}

	mux := http.NewServeMux()

	//estService := estserver.NewEstService(&lamassuCaClient, logger)

	//mux.Handle("/", lamassuest.MakeHTTPHandler(estService, log.With(logger, "component", "HTTPS"), tracer))
	http.Handle("/", accessControl(mux))
	mux.Handle("/", http.FileServer(http.Dir("./docs")))
	mux.Handle("/v1/", api.MakeHTTPHandler(s, log.With(logger, "component", "HTTPS"), tracer))
	mux.Handle("/v1/docs", middleware.SwaggerUI(middleware.SwaggerUIOpts{
		BasePath: "/v1",
		SpecURL:  path.Join("/openapiv3.json"),
		Path:     "docs",
	}, mux))

	http.Handle("/metrics", promhttp.Handler())

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		if strings.ToLower(cfg.Protocol) == "https" {
			if cfg.MutualTLSEnabled {
				mTlsCertPool, err := utils.CreateCAPool(cfg.MutualTLSClientCA)
				if err != nil {
					level.Error(logger).Log("err", err, "msg", "Could not create mTls Cert Pool")
					os.Exit(1)
				}
				tlsConfig := &tls.Config{
					ClientCAs:          mTlsCertPool,
					ClientAuth:         tls.RequireAndVerifyClientCert,
					InsecureSkipVerify: true,
				}

				tlsConfig.BuildNameToCertificate()

				http := &http.Server{
					Addr:      ":" + cfg.Port,
					TLSConfig: tlsConfig,
				}

				level.Info(logger).Log("transport", "Mutual TLS", "address", ":"+cfg.Port, "msg", "listening")
				errs <- http.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)

			} else {
				level.Info(logger).Log("transport", "HTTPS", "address", ":"+cfg.Port, "msg", "listening")
				errs <- http.ListenAndServeTLS(":"+cfg.Port, cfg.CertFile, cfg.KeyFile, nil)

			}
		} else if strings.ToLower(cfg.Protocol) == "http" {
			level.Info(logger).Log("transport", "HTTP", "address", ":"+cfg.Port, "msg", "listening")
			errs <- http.ListenAndServe(":"+cfg.Port, nil)

		} else {
			level.Error(logger).Log("err", "msg", "Unknown protocol")
			os.Exit(1)

		}
	}()
	level.Info(logger).Log("exit", <-errs)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
