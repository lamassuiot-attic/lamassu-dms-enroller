package api

import (
	"context"
	"enroller/pkg/enroller/crypto"
	"enroller/pkg/enroller/models/files"
	"errors"
	"os"
	"strconv"
	"sync"

	"enroller/pkg/enroller/models/db"
)

type Service interface {
	PostCSR(ctx context.Context, data []byte) (crypto.CSR, error)
	GetPendingCSRs(ctx context.Context) crypto.CSRs
	GetPendingCSRDB(ctx context.Context, id int) (crypto.CSR, error)
	GetPendingCSRFile(ctx context.Context, id int) ([]byte, error)
	PutChangeCSRStatus(ctx context.Context, csr crypto.CSR, id int) error
	DeleteCSR(ctx context.Context, id int) error
	GetCRT(ctx context.Context, id int) ([]byte, error)
}

type enrollerService struct {
	mtx          sync.RWMutex
	enrollerDB   db.DBCSRStore
	enrollerFile files.FileCSRStore
}

var (
	ErrIncorrectType    = errors.New("Incorrect media type")
	ErrIncorrectContent = errors.New("Incorrect data content")
	ErrBadRouting       = errors.New("Bad routing")
	ErrBadKey           = errors.New("Unexpected JWT key signing method")
)

func NewEnrollerService(enrollerDB db.DBCSRStore, file files.FileCSRStore) Service {
	return &enrollerService{
		enrollerDB:   enrollerDB,
		enrollerFile: file,
	}
}

func (s *enrollerService) PostCSR(ctx context.Context, data []byte) (crypto.CSR, error) {
	csr, err := crypto.ParseNewCSR(data)
	if err != nil {
		return crypto.CSR{}, ErrIncorrectContent
	}
	id, err := s.enrollerDB.InsertCSR(csr)
	if err != nil {
		return crypto.CSR{}, ErrIncorrectContent
	}
	csr.Id = id
	csr.CsrFilePath = os.Getenv("ENROLLER_HOME") + "/" + strconv.Itoa(id) + ".csr"
	err = s.enrollerDB.UpdateCSRFilePath(csr)
	if err != nil {
		return crypto.CSR{}, ErrIncorrectContent
	}
	err = s.enrollerFile.InsertFileCSR(id, data)
	if err != nil {
		s.enrollerDB.UpdateCSRByID(id, crypto.CSR{Status: crypto.DeniedStatus})
		return crypto.CSR{}, ErrIncorrectContent
	}

	return csr, nil
}

func (s *enrollerService) GetPendingCSRs(ctx context.Context) crypto.CSRs {
	csrs := s.enrollerDB.SelectCSRsByStatus(crypto.PendingStatus)
	return csrs
}

func (s *enrollerService) GetPendingCSRDB(ctx context.Context, id int) (crypto.CSR, error) {
	csr, err := s.enrollerDB.SelectCSRByID(id)
	if err != nil {
		return crypto.CSR{}, ErrIncorrectContent
	}
	return csr, nil
}

func (s *enrollerService) GetPendingCSRFile(ctx context.Context, id int) ([]byte, error) {
	data, err := s.enrollerFile.SelectFileByID(id)
	if err != nil {
		return nil, ErrIncorrectContent
	}
	return data, nil
}

func (s *enrollerService) signCSR(csr crypto.CSR, id int) error {
	csrData, err := s.enrollerFile.SelectFileByID(id)
	if err != nil {
		return ErrBadRouting
	}
	serial, err := s.enrollerFile.Serial()
	if err != nil {
		return ErrIncorrectContent
	}
	caCertData, err := s.enrollerFile.LoadCACert()
	if err != nil {
		return ErrBadRouting
	}
	caKeyData, err := s.enrollerFile.LoadCAKey()
	if err != nil {
		return ErrBadRouting
	}
	crtData, err := csr.SignCSR(csrData, serial, caCertData, caKeyData)
	if err != nil {
		return ErrIncorrectContent
	}
	err = s.enrollerFile.InsertFileCert(id, crtData)
	if err != nil {
		return ErrIncorrectContent
	}
	_, err = s.enrollerDB.UpdateCSRByID(id, csr)
	if err != nil {
		return ErrIncorrectContent
	}
	return nil
}

func (s *enrollerService) PutChangeCSRStatus(ctx context.Context, csr crypto.CSR, id int) error {
	var err error
	switch status := csr.Status; status {
	case crypto.ApprobedStatus:
		err = s.signCSR(csr, id)
	case crypto.DeniedStatus, crypto.PendingStatus:
		_, err = s.enrollerDB.UpdateCSRByID(id, csr)
	default:
		err = ErrIncorrectContent

	}
	return err
}

func (s *enrollerService) DeleteCSR(ctx context.Context, id int) error {
	_, err := s.enrollerDB.UpdateCSRByID(id, crypto.CSR{Status: crypto.DeniedStatus})
	if err != nil {
		return ErrIncorrectContent
	}
	err = s.enrollerFile.DeleteFile(id)
	if err != nil {
		return ErrIncorrectContent
	}
	return nil
}

func (s *enrollerService) GetCRT(ctx context.Context, id int) ([]byte, error) {
	data, err := s.enrollerFile.LoadCert(id)
	if err != nil {
		return nil, ErrBadRouting
	}

	return data, nil
}
