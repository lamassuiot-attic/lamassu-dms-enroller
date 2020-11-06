package main

import (
	"enroller/pkg/ca/api"
	"enroller/pkg/ca/auth"
	"enroller/pkg/ca/configs"
	"enroller/pkg/ca/secrets/vault"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
)

func main() {
	err, cfg := configs.NewConfig()
	if err != nil {
		panic(err)
	}

	auth := auth.NewAuth(cfg.KeycloakHostname, cfg.KeycloakPort, cfg.KeycloakProtocol, cfg.KeycloakRealm)
	secrets, err := vault.NewVaultSecrets(cfg.VaultAddress, cfg.VaultRoleID, cfg.VaultSecretID)

	if err != nil {
		panic(err)
	}

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

	var s api.Service
	{
		s = api.NewCAService(secrets)
		s = api.LoggingMiddleware(logger)(s)
	}

	mux := http.NewServeMux()

	mux.Handle("/v1/", api.MakeHTTPHandler(s, log.With(logger, "component", "HTTPS"), auth))
	http.Handle("/", accessControl(mux, cfg.EnrollerUIProtocol, cfg.EnrollerUIHost, cfg.EnrollerUIPort))

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTPS", "addr", "httpsAddr")
		errs <- http.ListenAndServeTLS(*httpAddr, cfg.CertFile, cfg.KeyFile, nil)
	}()

	logger.Log("exit", <-errs)

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
