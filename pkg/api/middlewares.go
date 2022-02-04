package api

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/lamassuiot/dms-enroller/pkg/models/dms"
	"github.com/opentracing/opentracing-go"

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
			"trace_id", opentracing.SpanFromContext(ctx),
		)
	}(time.Now())
	return mw.next.Health(ctx)
}

func (mw loggingMiddleware) CreateDMS(ctx context.Context, csrBase64Encoded string, dmsName string) (dms dms.DMS, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "CreateDMS",
			"csrBase64Encoded", csrBase64Encoded,
			"dmsName", dmsName,
			"dms", dms,
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
			"err", err,
		)
	}(time.Now())
	return mw.next.CreateDMS(ctx, csrBase64Encoded, dmsName)
}

func (mw loggingMiddleware) CreateDMSForm(ctx context.Context, dmsForm dms.DmsCreationForm) (_ string, d dms.DMS, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "CreateDMSForm",
			"dmsForm", dmsForm,
			"dms", d,
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
			"err", err,
		)
	}(time.Now())
	return mw.next.CreateDMSForm(ctx, dmsForm)
}

func (mw loggingMiddleware) UpdateDMSStatus(ctx context.Context, dIn dms.DMS, id int) (dOut dms.DMS, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "UpdateDMSStatus",
			"id", id,
			"dms_in", dIn,
			"dms_out", dOut,
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
		)
	}(time.Now())
	return mw.next.UpdateDMSStatus(ctx, dIn, id)
}

func (mw loggingMiddleware) DeleteDMS(ctx context.Context, id int) (err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "DeleteDMS",
			"id", id,
			"err", err,
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
		)
	}(time.Now())
	return mw.next.DeleteDMS(ctx, id)
}

func (mw loggingMiddleware) GetDMSCertificate(ctx context.Context, id int) (crt *x509.Certificate, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetDMSCertificate",
			"id", id,
			"crt_serialnumber", crt.SerialNumber.String(),
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
			"err", err,
		)
	}(time.Now())
	return mw.next.GetDMSCertificate(ctx, id)
}

func (mw loggingMiddleware) GetDMSs(ctx context.Context) (d []dms.DMS, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"method", "GetDMSs",
			"dmss", d,
			"took", time.Since(begin),
			"trace_id", opentracing.SpanFromContext(ctx),
			"err", err,
		)
	}(time.Now())
	return mw.next.GetDMSs(ctx)
}
