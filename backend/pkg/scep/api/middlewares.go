package api

import (
	"context"
	"enroller/pkg/scep/crypto"
	"time"

	"github.com/go-kit/kit/log"
)

type Middleware func(Service) Service

func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) Health(ctx context.Context) (healthy bool) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "Health",
			"took", time.Since(begin),
			"healthy", healthy,
		)
	}(time.Now())
	return mw.next.Health(ctx)
}

func (mw loggingMiddleware) GetSCEPCRTs(ctx context.Context) (crts crypto.CRTs, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetSCEPCRTs",
			"number_crts", len(crts.CRTs),
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.GetSCEPCRTs(ctx)
}

func (mw loggingMiddleware) RevokeSCEPCRT(ctx context.Context, dn string, serial string) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "RevokeSCEPCRT",
			"dn", dn,
			"serial", serial,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.RevokeSCEPCRT(ctx, dn, serial)
}
