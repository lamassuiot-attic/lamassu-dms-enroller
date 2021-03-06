package api

import (
	"context"
	"fmt"
	"time"

	"github.com/lamassuiot/enroller/pkg/scep/crypto"

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

func (mw *instrumentingMiddleware) Health(ctx context.Context) bool {
	defer func(begin time.Time) {
		lvs := []string{"method", "Health", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Health(ctx)
}

func (mw *instrumentingMiddleware) GetSCEPCRTs(ctx context.Context) (crts crypto.CRTs, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetSCEPCRTs", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetSCEPCRTs(ctx)
}

func (mw *instrumentingMiddleware) RevokeSCEPCRT(ctx context.Context, dn string, serial string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RevokeSCEPCRT", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.RevokeSCEPCRT(ctx, dn, serial)
}
