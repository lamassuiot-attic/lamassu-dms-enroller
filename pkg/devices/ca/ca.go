package ca

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/globalsign/est"
	"github.com/go-kit/kit/log"
	"github.com/lamassuiot/lamassu-est/client/estclient"

	devicesModel "github.com/lamassuiot/enroller/pkg/devices/models/device"
	devicesStore "github.com/lamassuiot/enroller/pkg/devices/models/device/store"
)

type DeviceService struct {
	devicesDb devicesStore.DB
	logger    log.Logger
}

func NewVaultService(devicesDb devicesStore.DB) *DeviceService {
	return &DeviceService{
		devicesDb: devicesDb,
	}
}

func (ca *DeviceService) CACerts(ctx context.Context, aps string, req *http.Request) ([]*x509.Certificate, error) {

	var filteredCerts []*x509.Certificate

	return filteredCerts, nil
}

func (ca *DeviceService) Enroll(ctx context.Context, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, error) {
	deviceId := csr.Subject.CommonName
	device, err := ca.devicesDb.SelectDeviceById(deviceId)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			fmt.Println("Device " + deviceId + " does not exist. Register the device first, and enroll it afterwards")
		}
		return nil, err
	}
	if device.Status == devicesModel.DeviceDecommisioned {
		err := "Cant issue a certificate for a decommisioned device"
		fmt.Println(err)
		return nil, errors.New(err)
	}
	if device.Status == devicesModel.DeviceProvisioned {
		err := "The device (" + deviceId + ") already has a valid certificate"
		fmt.Println(err)
		return nil, errors.New(err)
	}

	cert, err := estclient.Enroll(csr, aps)
	if err != nil {
		return nil, err
	}

	deviceId = cert.Subject.CommonName
	fmt.Println("Device ID: " + deviceId)
	log := devicesModel.DeviceLog{
		DeviceId:   deviceId,
		LogType:    devicesModel.LogProvisioned,
		LogMessage: "",
	}
	err = ca.devicesDb.InsertLog(log)
	if err != nil {
		return nil, err
	}

	serialNumber := insertNth(toHexInt(cert.SerialNumber), 2)
	certHistory := devicesModel.DeviceCertHistory{
		SerialNumber: serialNumber,
		DeviceId:     deviceId,
		IsuuerName:   aps,
		Status:       devicesModel.CertHistoryActive,
	}
	err = ca.devicesDb.InsertDeviceCertHistory(certHistory)
	if err != nil {
		return nil, err
	}

	err = ca.devicesDb.UpdateDeviceStatusByID(deviceId, devicesModel.DeviceProvisioned)
	if err != nil {
		return nil, err
	}

	err = ca.devicesDb.UpdateDeviceCertificateSerialNumberByID(deviceId, serialNumber)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (ca *DeviceService) CSRAttrs(ctx context.Context, aps string, r *http.Request) (est.CSRAttrs, error) {
	return est.CSRAttrs{}, nil
}

func (ca *DeviceService) Reenroll(ctx context.Context, cert *x509.Certificate, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, error) {
	newCert, err := estclient.Reenroll(csr, aps)
	if err != nil {
		return nil, err
	}
	return newCert, nil
}

func (ca *DeviceService) ServerKeyGen(ctx context.Context, csr *x509.CertificateRequest, aps string, r *http.Request) (*x509.Certificate, []byte, error) {
	return nil, nil, nil
}

func (ca *DeviceService) TPMEnroll(ctx context.Context, csr *x509.CertificateRequest, ekcerts []*x509.Certificate, ekPub, akPub []byte, aps string, r *http.Request) ([]byte, []byte, []byte, error) {
	return nil, nil, nil, nil
}

func toHexInt(n *big.Int) string {
	return fmt.Sprintf("%x", n) // or %X or upper case
}

func insertNth(s string, n int) string {
	if len(s)%2 != 0 {
		s = "0" + s
	}
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, rune := range s {
		buffer.WriteRune(rune)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('-')
		}
	}
	return buffer.String()
}
