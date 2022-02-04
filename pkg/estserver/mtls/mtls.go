package mtls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	stdhttp "net/http"

	"errors"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/http"
)

type contextKey string

const (
	PeerCertificatesContextKey contextKey = "PeerCertificatesContextKey"
)

var (
	ErrPeerCertificatesContextMissing = errors.New("token up for parsing was not passed through the context")
)

func HTTPToContext() http.RequestFunc {
	return func(ctx context.Context, r *stdhttp.Request) context.Context {
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			return context.WithValue(ctx, PeerCertificatesContextKey, r.TLS.PeerCertificates[0])
		} else {
			return ctx
		}
	}
}

//Configurar la función Verify para verificar que el certificado del cliente ha sido firmado por una CA de confianza, sino sacar un errror
func NewParser() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			peerCert, ok := ctx.Value(PeerCertificatesContextKey).(*x509.Certificate)
			if !ok {
				return nil, ErrPeerCertificatesContextMissing
			}

			pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: peerCert.Raw})

			fmt.Println(string(pemCert))

			return next(ctx, request)
		}
	}
}
