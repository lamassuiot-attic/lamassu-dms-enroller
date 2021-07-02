package ca

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"math/big"
	"net/http"

	"github.com/globalsign/est"
	"github.com/lamassuiot/lamassu-est/client/estclient"

	devicesModel "github.com/lamassuiot/enroller/pkg/devices/models/device"
	devicesStore "github.com/lamassuiot/enroller/pkg/devices/models/device/store"
)

type DeviceService struct {
	devicesDb devicesStore.DB
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
	cert, err := estclient.Enroll(csr, aps)
	if err != nil {
		return nil, err
	}

	log := devicesModel.DeviceLog{
		DeviceId:   "c9eddb42-558d-4371-9f91-90ef29ad768d",
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
		DeviceId:     "c9eddb42-558d-4371-9f91-90ef29ad768d",
		IsuuerName:   aps,
		Status:       devicesModel.CertHistoryActive,
	}
	err = ca.devicesDb.InsertDeviceCertHistory(certHistory)
	if err != nil {
		return nil, err
	}

	err = ca.devicesDb.UpdateDeviceStatusByID("c9eddb42-558d-4371-9f91-90ef29ad768d", devicesModel.DeviceProvisioned)
	if err != nil {
		return nil, err
	}

	err = ca.devicesDb.UpdateDeviceCertificateSerialNumberByID("c9eddb42-558d-4371-9f91-90ef29ad768d", serialNumber)
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
