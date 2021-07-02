package store

import "github.com/lamassuiot/enroller/pkg/devices/models/device"

type DB interface {
	InsertDevice(d device.Device) error
	SelectDeviceById(id string) (device.Device, error)
	SelectAllDevices() (device.Devices, error)
	SelectAllDevicesByDmsId(dms_id string) (device.Devices, error)
	UpdateDeviceStatusByID(id string, newStatus string) error
	UpdateDeviceCertificateSerialNumberByID(id string, serialNumber string) error
	DeleteDevice(id string) error

	InsertLog(l device.DeviceLog) error
	SelectDeviceLogs(id string) (device.DeviceLogs, error)

	InsertDeviceCertHistory(l device.DeviceCertHistory) error
	SelectDeviceCertHistory(deviceId string) (device.DeviceCertsHistory, error)
	SelectDeviceCertHistoryBySerialNumber(serialNumber string) (device.DeviceCertHistory, error)
	UpdateDeviceCertHistory(deviceId string, serialNumber string, newStatus string) error
}
