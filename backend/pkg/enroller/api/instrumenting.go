package api

import (
	"context"
	csrmodel "enroller/pkg/enroller/models/csr"
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

func (mw *instrumentingMiddleware) PostCSR(ctx context.Context, data []byte) (csrmodel.CSR, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "PostCSR").Add(1)
		mw.requestLatency.With("method", "PostCSR").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.PostCSR(ctx, data)
}

func (mw *instrumentingMiddleware) GetPendingCSRs(ctx context.Context) csrmodel.CSRs {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "GetPendingCSRs").Add(1)
		mw.requestLatency.With("method", "GetPendingCSRs").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRs(ctx)
}

func (mw *instrumentingMiddleware) GetPendingCSRDB(ctx context.Context, id int) (csrmodel.CSR, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "GetPendingCSRDB").Add(1)
		mw.requestLatency.With("method", "GetPendingCSRDB").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRDB(ctx, id)
}

func (mw *instrumentingMiddleware) GetPendingCSRFile(ctx context.Context, id int) ([]byte, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "GetPendingCSRFile").Add(1)
		mw.requestLatency.With("method", "GetPendingCSRFile").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRFile(ctx, id)
}

func (mw *instrumentingMiddleware) PutChangeCSRStatus(ctx context.Context, csr csrmodel.CSR, id int) (csrmodel.CSR, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "PutChangeCSRStatus").Add(1)
		mw.requestLatency.With("method", "PutChangeCSRStatus").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.PutChangeCSRStatus(ctx, csr, id)
}

func (mw *instrumentingMiddleware) DeleteCSR(ctx context.Context, id int) error {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "DeleteCSR").Add(1)
		mw.requestLatency.With("method", "DeleteCSR").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.DeleteCSR(ctx, id)
}

func (mw *instrumentingMiddleware) GetCRT(ctx context.Context, id int) ([]byte, error) {
	defer func(begin time.Time) {
		mw.requestCount.With("method", "GetCRT").Add(1)
		mw.requestLatency.With("method", "GetCRT").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetCRT(ctx, id)
}
