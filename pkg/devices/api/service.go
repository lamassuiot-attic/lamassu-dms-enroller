package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	devicesModel "github.com/lamassuiot/enroller/pkg/devices/models/device"
	devicesStore "github.com/lamassuiot/enroller/pkg/devices/models/device/store"
)

type Service interface {
	Health(ctx context.Context) bool
	PostDevice(ctx context.Context, device devicesModel.Device) (devicesModel.Device, error)
	GetDevices(ctx context.Context) (devicesModel.Devices, error)
	GetDeviceById(ctx context.Context, deviceId string) (devicesModel.Device, error)
	GetDevicesByDMS(ctx context.Context, dmsId string) (devicesModel.Devices, error)
	DeleteDevice(ctx context.Context, id string) error
	IssueDeviceCert(ctx context.Context, id string) (string, error)
	RevokeDeviceCert(ctx context.Context, id string) error

	GetDeviceLogs(ctx context.Context, id string) (devicesModel.DeviceLogs, error)
	GetDeviceCertHistory(ctx context.Context, id string) (devicesModel.DeviceCertsHistory, error)
}

type devicesService struct {
	mtx       sync.RWMutex
	devicesDb devicesStore.DB
}

var (
	// Client errors
	ErrInvalidDeviceRequest = errors.New("unable to parse device, is invalid")    //400
	ErrInvalidDMSId         = errors.New("unable to parse DMS ID, is invalid")    //400
	ErrInvalidDeviceId      = errors.New("unable to parse Device ID, is invalid") //400
	ErrIncorrectType        = errors.New("unsupported media type")                //415
	ErrEmptyBody            = errors.New("empty body")

	//Server errors
	ErrInvalidOperation = errors.New("invalid operation")
	ErrActiveCert       = errors.New("can't isuee certificate. The device has a valid cert")
	ErrResponseEncode   = errors.New("error encoding response")
)

func NewDevicesService(devicesDb devicesStore.DB) Service {
	return &devicesService{
		devicesDb: devicesDb,
	}
}

func (s *devicesService) Health(ctx context.Context) bool {
	return true
}

func (s *devicesService) PostDevice(ctx context.Context, device devicesModel.Device) (devicesModel.Device, error) {
	device.KeyStrength = getKeyStrength(device.KeyType, device.KeyBits)
	err := s.devicesDb.InsertDevice(device)
	if err != nil {
		return devicesModel.Device{}, err
	}

	log := devicesModel.DeviceLog{
		DeviceId:   device.Id,
		LogType:    devicesModel.LogDeviceCreated,
		LogMessage: "",
	}
	err = s.devicesDb.InsertLog(log)
	if err != nil {
		return devicesModel.Device{}, err
	}
	log = devicesModel.DeviceLog{
		DeviceId:   device.Id,
		LogType:    devicesModel.LogPendingProvision,
		LogMessage: "",
	}
	err = s.devicesDb.InsertLog(log)
	if err != nil {
		return devicesModel.Device{}, err
	}

	device, err = s.devicesDb.SelectDeviceById(device.Id)
	if err != nil {
		return devicesModel.Device{}, err
	}
	return device, nil
}

func (s *devicesService) GetDevices(ctx context.Context) (devicesModel.Devices, error) {
	devices, err := s.devicesDb.SelectAllDevices()
	if err != nil {
		return devicesModel.Devices{}, err
	}

	return devices, nil
}

func (s *devicesService) GetDevicesByDMS(ctx context.Context, dmsId string) (devicesModel.Devices, error) {
	devices, err := s.devicesDb.SelectAllDevicesByDmsId(dmsId)
	if err != nil {
		return devicesModel.Devices{}, err
	}

	return devices, nil
}
func (s *devicesService) GetDeviceById(ctx context.Context, deviceId string) (devicesModel.Device, error) {
	device, err := s.devicesDb.SelectDeviceById(deviceId)
	if err != nil {
		return devicesModel.Device{}, err
	}

	return device, nil
}

func (s *devicesService) DeleteDevice(ctx context.Context, id string) error {
	/*err := s.devicesDb.DeleteDevice(id)
	if err != nil {
		return err
	}*/

	err := s.devicesDb.UpdateDeviceStatusByID(id, devicesModel.DeviceDecomissioned)
	if err != nil {
		return err
	}

	log := devicesModel.DeviceLog{
		DeviceId:   id,
		LogType:    devicesModel.LogDeviceDecomissioned,
		LogMessage: "",
	}
	err = s.devicesDb.InsertLog(log)
	if err != nil {
		return err
	}
	return err
}

func (s *devicesService) IssueDeviceCert(ctx context.Context, id string) (string, error) {
	// GET LAST CERT ID & If revoked, update status to "Revoked" else return error: CANT ISSUE
	currentCertHistory, err := s.devicesDb.SelectDeviceLastCertHistory(id)
	if err != nil {
		return "", err
	}
	var empty devicesModel.DeviceCertHistory
	if currentCertHistory == empty {
		fmt.Println("Empty")
	} else if currentCertHistory.Status == devicesModel.CertHistoryActive {
		fmt.Println("TODO: Check if cert is expired")

		// Check if cert is expired

		// -> Is not Expired
		//		return "", ErrActiveCert
		// -> Is Expired
		//		OK: Issue cert

		// ISSUE VAULT CERT (debera devolver 4 cosas: crt, crt-serial, crt-issuer-serial, crt-issuer-name(aka. CA) )
	}

	// http client
	/*postBody, _ := json.Marshal(map[string]string{
		"name":  "Toby",
		"email": "Toby@example.com",
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("http://localhost:8080/", "application/json", responseBody)
	if err != nil {
		return "", err
	}*/

	err = s.devicesDb.UpdateDeviceLastCertHistory(id, device.CertHistoryRevoked)
	if err != nil {
		return "", err
	}

	log := devicesModel.DeviceLog{
		DeviceId:   id,
		LogType:    devicesModel.LogProvisioned,
		LogMessage: "",
	}
	err = s.devicesDb.InsertLog(log)
	if err != nil {
		return "", err
	}

	err = s.devicesDb.UpdateDeviceStatusByID(id, devicesModel.DeviceProvisioned)
	if err != nil  {
		return "", err
	}

	certHistory := devicesModel.DeviceCertHistory{
		DeviceId:           id,
		SerialNumber:       "",
		IssuerSerialNumber: "",
		IsuuerName:         "",
		Status:             devicesModel.CertHistoryActive,
	}
	s.devicesDb.InsertDeviceCertHistory(certHistory)

	//	DeviceProvisioned
	if err != nil {
		return "", err
	}

	return "", nil
}

func (s *devicesService) RevokeDeviceCert(ctx context.Context, id string) error {
	/*currentCertHistory, err := s.devicesDb.SelectDeviceLastCertHistory(id)
	if err != nil {
		return err
	}*/
	// revoke
	fmt.Println("TODO: Revoke cert ")
	/*err = s.devicesDb.UpdateDeviceLastCertHistory(id, devicesModel.CertHistoryRevoked)
	if err != nil {
		return err
	}*/

	err := s.devicesDb.UpdateDeviceStatusByID(id, devicesModel.DeviceCertRevoked)
	if err != nil {
		return err
	}

	log := devicesModel.DeviceLog{
		DeviceId:   id,
		LogType:    devicesModel.LogCertRevoked,
		LogMessage: "",
	}
	err = s.devicesDb.InsertLog(log)
	if err != nil {
		return err
	}
	return nil
}

func (s *devicesService) GetDeviceLogs(ctx context.Context, id string) (devicesModel.DeviceLogs, error) {
	logs, err := s.devicesDb.SelectDeviceLogs(id)
	if err != nil {
		return devicesModel.DeviceLogs{}, err
	}
	return logs, nil
}

func (s *devicesService) GetDeviceCertHistory(ctx context.Context, id string) (devicesModel.DeviceCertsHistory, error) {
	history, err := s.devicesDb.SelectDeviceCertHistory(id)
	if err != nil {
		return devicesModel.DeviceCertsHistory{}, err
	}
	return history, nil
}

func getKeyStrength(keyType string, keyBits int) string {
	var keyStrength string = "unknown"
	switch keyType {
	case "rsa":
		if keyBits < 2048 {
			keyStrength = "low"
		} else if keyBits >= 2048 && keyBits < 3072 {
			keyStrength = "medium"
		} else {
			keyStrength = "high"
		}
	case "ec":
		if keyBits < 224 {
			keyStrength = "low"
		} else if keyBits >= 224 && keyBits < 256 {
			keyStrength = "medium"
		} else {
			keyStrength = "high"
		}
	}
	return keyStrength
}
