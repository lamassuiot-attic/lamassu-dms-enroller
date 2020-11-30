package api

import (
	"context"
	"enroller/pkg/scep/crypto"
	"time"

	"github.com/go-kit/kit/metrics"
)

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

func NewInstrumentingMiddleware(counter metrics.Counter, latency metrics.Histogram) Middleware {
	return func(next Service) Service {
		return &instrumentingMiddleware{
			requestCount:   counter,
			requestLatency: latency,
			next:           next,
		}
	}
}
func (mw *instrumentingMiddleware) GetSCEPCRTs(ctx context.Context) (crts crypto.CRTs, err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "GetSCEPCRTs").Add(1)
		mw.requestLatency.With("method", "GetSCEPCRTs").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetSCEPCRTs(ctx)
}

func (mw *instrumentingMiddleware) RevokeSCEPCRT(ctx context.Context, dn string, serial string) (err error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "RevokeSCEPCRT").Add(1)
		mw.requestLatency.With("method", "RevokeSCEPCRT").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.RevokeSCEPCRT(ctx, dn, serial)
}
