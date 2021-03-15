package api

import (
	"context"
	csrmodel "enroller/pkg/enroller/models/csr"
	"fmt"
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

func (mw *instrumentingMiddleware) Health(ctx context.Context) bool {
	defer func(begin time.Time) {
		lvs := []string{"method", "Health", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.Health(ctx)
}

func (mw *instrumentingMiddleware) PostCSR(ctx context.Context, data []byte) (csr csrmodel.CSR, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "PostCSR", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.PostCSR(ctx, data)
}

func (mw *instrumentingMiddleware) GetPendingCSRs(ctx context.Context) csrmodel.CSRs {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetPendingCSRs", "error", "false"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRs(ctx)
}

func (mw *instrumentingMiddleware) GetPendingCSRDB(ctx context.Context, id int) (csr csrmodel.CSR, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetPendingCSRDB", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRDB(ctx, id)
}

func (mw *instrumentingMiddleware) GetPendingCSRFile(ctx context.Context, id int) (csr []byte, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetPendingCSRFile", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetPendingCSRFile(ctx, id)
}

func (mw *instrumentingMiddleware) PutChangeCSRStatus(ctx context.Context, csr csrmodel.CSR, id int) (c csrmodel.CSR, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "PutChangeCSRStatus", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.PutChangeCSRStatus(ctx, csr, id)
}

func (mw *instrumentingMiddleware) DeleteCSR(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "DeleteCSR", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.DeleteCSR(ctx, id)
}

func (mw *instrumentingMiddleware) GetCRT(ctx context.Context, id int) (crt []byte, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetCRT", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetCRT(ctx, id)
}
