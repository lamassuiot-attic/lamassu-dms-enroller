package api

import (
	"context"
	"enroller/pkg/enroller/crypto"
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

func (mw loggingMiddleware) PostCSR(ctx context.Context, data []byte) (csr crypto.CSR, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PostCSR", "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PostCSR(ctx, data)
}

func (mw loggingMiddleware) GetPendingCSRs(ctx context.Context) (csrs crypto.CSRs) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetPendingCSRs", "took", time.Since(begin))
	}(time.Now())
	return mw.next.GetPendingCSRs(ctx)
}

func (mw loggingMiddleware) GetPendingCSRDB(ctx context.Context, id int) (csr crypto.CSR, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetPendingCSRDB", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetPendingCSRDB(ctx, id)
}

func (mw loggingMiddleware) GetPendingCSRFile(ctx context.Context, id int) (data []byte, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetPendingCSRFile", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetPendingCSRFile(ctx, id)
}

func (mw loggingMiddleware) PutChangeCSRStatus(ctx context.Context, csr crypto.CSR, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "PutChangeCSRStatus", "id", id, "status", csr.Status, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.PutChangeCSRStatus(ctx, csr, id)
}

func (mw loggingMiddleware) DeleteCSR(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "DeleteCSR", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.DeleteCSR(ctx, id)
}

func (mw loggingMiddleware) GetCRT(ctx context.Context, id int) (data []byte, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "GetCRT", "id", id, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.GetCRT(ctx, id)
}
