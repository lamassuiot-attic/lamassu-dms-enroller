package api

import (
	"context"
	"time"

	"github.com/lamassuiot/enroller/pkg/enroller/models/csr"

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

func (mw loggingMiddleware) PostCSR(ctx context.Context, data []byte, dmsName string) (csr csr.CSR, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "PostCSR",
			"cn", csr.CommonName,
			"dmsName", dmsName,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.PostCSR(ctx, data, dmsName)
}

func (mw loggingMiddleware) GetPendingCSRs(ctx context.Context) (csrs csr.CSRs) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetPendingCSRs",
			"number_csrs", len(csrs.CSRs),
			"took", time.Since(begin),
		)
	}(time.Now())
	return mw.next.GetPendingCSRs(ctx)
}

func (mw loggingMiddleware) GetPendingCSRDB(ctx context.Context, id int) (csr csr.CSR, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetPendingCSRDB",
			"request_csr_id", id,
			"response_csr_id", csr.Id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.GetPendingCSRDB(ctx, id)
}

func (mw loggingMiddleware) GetPendingCSRFile(ctx context.Context, id int) (data []byte, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetPendingCSRFile",
			"request_csr_id", id,
			"took",
			time.Since(begin), "err", err,
		)
	}(time.Now())
	return mw.next.GetPendingCSRFile(ctx, id)
}

func (mw loggingMiddleware) PutChangeCSRStatus(ctx context.Context, csr csr.CSR, id int) (c csr.CSR, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "PutChangeCSRStatus",
			"request_csr_id", id,
			"request_csr_status", csr.Status,
			"response_csr_id", c.Id,
			"response_csr_status", c.Status,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.PutChangeCSRStatus(ctx, csr, id)
}

func (mw loggingMiddleware) DeleteCSR(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "DeleteCSR",
			"request_csr_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.DeleteCSR(ctx, id)
}

func (mw loggingMiddleware) GetCRT(ctx context.Context, id int) (data []byte, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetCRT",
			"request_crt_id", id,
			"took", time.Since(begin),
			"err", err,
		)
	}(time.Now())
	return mw.next.GetCRT(ctx, id)
}
