package api

import (
	"context"

	"github.com/lamassuiot/enroller/pkg/devices/models/device"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/tracing/opentracing"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint     endpoint.Endpoint
	PostDeviceEndpoint endpoint.Endpoint
	GetDevices         endpoint.Endpoint
	GetDeviceById      endpoint.Endpoint
	GetDevicesByDMS    endpoint.Endpoint
	DeleteDevice       endpoint.Endpoint
	PostIssue          endpoint.Endpoint
	DeleteRevoke       endpoint.Endpoint
	GetDeviceLogs      endpoint.Endpoint
}

func MakeServerEndpoints(s Service, otTracer stdopentracing.Tracer) Endpoints {
	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
		healthEndpoint = opentracing.TraceServer(otTracer, "Health")(healthEndpoint)
	}
	var postDeviceEndpoint endpoint.Endpoint
	{
		postDeviceEndpoint = MakePostDeviceEndpoint(s)
		postDeviceEndpoint = opentracing.TraceServer(otTracer, "PostCSR")(postDeviceEndpoint)
	}
	var getDevicesEndpoint endpoint.Endpoint
	{
		getDevicesEndpoint = MakeGetDevicesEndpoint(s)
		getDevicesEndpoint = opentracing.TraceServer(otTracer, "GetDevices")(getDevicesEndpoint)
	}
	var getDevicesByIdEndpoint endpoint.Endpoint
	{
		getDevicesByIdEndpoint = MakeGetDeviceByIdEndpoint(s)
		getDevicesByIdEndpoint = opentracing.TraceServer(otTracer, "GetDeviceById")(getDevicesByIdEndpoint)
	}
	var getDevicesByDMSEndpoint endpoint.Endpoint
	{
		getDevicesByDMSEndpoint = MakeGetDevicesByDMSEndpoint(s)
		getDevicesByDMSEndpoint = opentracing.TraceServer(otTracer, "GetDevicesByDMS")(getDevicesByDMSEndpoint)
	}
	var deleteDeviceEndpoint endpoint.Endpoint
	{
		deleteDeviceEndpoint = MakeDeleteDeviceEndpoint(s)
		deleteDeviceEndpoint = opentracing.TraceServer(otTracer, "DeleteDevice")(deleteDeviceEndpoint)
	}
	var postIssueEndpoint endpoint.Endpoint
	{
		postIssueEndpoint = MakePostIssueEndpoint(s)
		postIssueEndpoint = opentracing.TraceServer(otTracer, "PostIssue")(postIssueEndpoint)
	}
	var deleteRevokeEndpoint endpoint.Endpoint
	{
		deleteRevokeEndpoint = MakeDeleteRevokeEndpoint(s)
		deleteRevokeEndpoint = opentracing.TraceServer(otTracer, "deleteRevokeEndpoint")(deleteRevokeEndpoint)
	}
	var getDeviceLogsEndpoint endpoint.Endpoint
	{
		getDeviceLogsEndpoint = MakeGetDeviceLogsEndpoint(s)
		getDeviceLogsEndpoint = opentracing.TraceServer(otTracer, "getDeviceLogsEndpoint")(getDeviceLogsEndpoint)
	}

	return Endpoints{
		HealthEndpoint:     healthEndpoint,
		PostDeviceEndpoint: postDeviceEndpoint,
		GetDevices:         getDevicesEndpoint,
		GetDeviceById:      getDevicesByIdEndpoint,
		GetDevicesByDMS:    getDevicesByDMSEndpoint,
		DeleteDevice:       deleteDeviceEndpoint,
		PostIssue:          postIssueEndpoint,
		DeleteRevoke:       deleteRevokeEndpoint,
		GetDeviceLogs:      getDeviceLogsEndpoint,
	}
}

func MakeHealthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return healthResponse{Healthy: healthy}, nil
	}
}

func MakePostDeviceEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postDeviceRequest)
		device, e := s.PostDevice(ctx, req.Device)
		return postDeviceResponse{Device: device, Err: e}, nil
	}
}

func MakeGetDevicesEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		devices, e := s.GetDevices(ctx)
		return devices.Devices, e
	}
}
func MakeGetDeviceByIdEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getDevicesByIdRequest)
		device, e := s.GetDeviceById(ctx, req.Id)
		return device, e
	}
}
func MakeGetDevicesByDMSEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getDevicesByDMSRequest)
		devices, e := s.GetDevicesByDMS(ctx, req.Id)
		return devices.Devices, e
	}
}
func MakeDeleteDeviceEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteDeviceRequest)
		e := s.DeleteDevice(ctx, req.Id)
		if e != nil {
			return "", e
		} else {
			return "OK", e
		}
	}
}
func MakePostIssueEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(postIssueRequest)
		cert, e := s.IssueDeviceCert(ctx, req.Id)
		return cert, e
	}
}
func MakeDeleteRevokeEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(deleteRevokeRequest)
		e := s.RevokeDeviceCert(ctx, req.Id)
		return nil, e
	}
}
func MakeGetDeviceLogsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(getDeviceLogsRequest)
		logs, e := s.GetDeviceLogs(ctx, req.Id)
		return logs.Logs, e
	}
}

type healthRequest struct{}

type healthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"err,omitempty"`
}

type postDeviceRequest struct {
	Device device.Device
}

type postDeviceResponse struct {
	Device device.Device `json:"device,omitempty"`
	Err    error         `json:"err,omitempty"`
}

func (r postDeviceResponse) error() error { return r.Err }

type getDevicesResponse struct {
	Devices []device.Device `json:"devices,omitempty"`
	Err     error           `json:"err,omitempty"`
}

type getDevicesByIdRequest struct {
	Id string
}
type getDevicesByDMSRequest struct {
	Id string
}
type deleteDeviceRequest struct {
	Id string
}
type postIssueRequest struct {
	Id string
}
type deleteRevokeRequest struct {
	Id string
}
type getDeviceLogsRequest struct {
	Id string
}
