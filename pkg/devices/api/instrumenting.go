package api

import (
	"context"
	"fmt"
	"time"

	deviceModel "github.com/lamassuiot/enroller/pkg/devices/models/device"
	devicesModel "github.com/lamassuiot/enroller/pkg/devices/models/device"

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

func (mw *instrumentingMiddleware) PostDevice(ctx context.Context, device deviceModel.Device) (deviceResp deviceModel.Device, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "PostDevice", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.PostDevice(ctx, device)
}

func (mw *instrumentingMiddleware) GetDevices(ctx context.Context) (device devicesModel.Devices, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetDevices", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetDevices(ctx)
}

func (mw *instrumentingMiddleware) GetDeviceById(ctx context.Context, deviceId string) (device devicesModel.Device, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetDeviceById", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetDeviceById(ctx, deviceId)
}

func (mw *instrumentingMiddleware) GetDevicesByDMS(ctx context.Context, dmsId string) (devices devicesModel.Devices, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetDevicesByDMS", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetDevicesByDMS(ctx, dmsId)
}
func (mw *instrumentingMiddleware) DeleteDevice(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "DeleteDevice", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.DeleteDevice(ctx, id)
}
func (mw *instrumentingMiddleware) IssueDeviceCert(ctx context.Context, id string, csr []byte) (cert string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "IssueDeviceCert", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.IssueDeviceCert(ctx, id, csr)
}
func (mw *instrumentingMiddleware) IssueDeviceCertUsingDefaults(ctx context.Context, id string) (privKey string, cert string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "IssueDeviceCertUsingDefaults", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.IssueDeviceCertUsingDefaults(ctx, id)
}
func (mw *instrumentingMiddleware) RevokeDeviceCert(ctx context.Context, id string) (err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RevokeDeviceCert", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.RevokeDeviceCert(ctx, id)
}
func (mw *instrumentingMiddleware) GetDeviceLogs(ctx context.Context, id string) (logs devicesModel.DeviceLogs, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetDeviceLogs", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetDeviceLogs(ctx, id)
}
func (mw *instrumentingMiddleware) GetDeviceCertHistory(ctx context.Context, id string) (history devicesModel.DeviceCertsHistory, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetDeviceCertHistory", "error", fmt.Sprint(err != nil)}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mw.next.GetDeviceCertHistory(ctx, id)
}
