package api

import (
	"bytes"
	"context"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lamassuiot/enroller/pkg/enroller/auth"
	"github.com/lamassuiot/enroller/pkg/enroller/crypto"
	"github.com/lamassuiot/enroller/pkg/enroller/models/certs"
	certstore "github.com/lamassuiot/enroller/pkg/enroller/models/certs/store"
	"github.com/lamassuiot/enroller/pkg/enroller/models/csr"
	csrmodel "github.com/lamassuiot/enroller/pkg/enroller/models/csr"
	csrstore "github.com/lamassuiot/enroller/pkg/enroller/models/csr/store"
	"github.com/lamassuiot/enroller/pkg/enroller/secrets"

	"github.com/go-kit/kit/auth/jwt"
)

type Service interface {
	Health(ctx context.Context) bool
	PostCSR(ctx context.Context, data []byte) (csrmodel.CSR, error)
	GetPendingCSRs(ctx context.Context) csrmodel.CSRs
	GetPendingCSRDB(ctx context.Context, id int) (csrmodel.CSR, error)
	GetPendingCSRFile(ctx context.Context, id int) ([]byte, error)
	PutChangeCSRStatus(ctx context.Context, csr csrmodel.CSR, id int) (csrmodel.CSR, error)
	DeleteCSR(ctx context.Context, id int) error
	GetCRT(ctx context.Context, id int) ([]byte, error)
}

type enrollerService struct {
	mtx            sync.RWMutex
	csrDBStore     csrstore.DB
	csrFileStore   csrstore.File
	certsDBStore   certstore.DB
	certsFileStore certstore.File
	secrets        secrets.Secrets
	homePath       string
}

var (
	// Client errors
	ErrInvalidCSR       = errors.New("unable to parse CSR, is invalid") //400
	ErrInvalidID        = errors.New("invalid CSR ID, does not exist")  //404
	ErrInvalidIDFormat  = errors.New("invalid ID format")
	ErrInvalidApprobeOp = errors.New("invalid operation, only pending status CSRs can be approved")          //400
	ErrInvalidRevokeOp  = errors.New("invalid operation, only approved status CSRs can be revoked")          //400
	ErrInvalidDenyOp    = errors.New("invalid operation, only pending status CSRs can be denied")            //400
	ErrInvalidDeleteOp  = errors.New("invalid operation, only denied or revoked status CSRs can be deleted") //400
	ErrIncorrectType    = errors.New("unsupported media type")                                               //415
	ErrEmptyBody        = errors.New("empty body")

	//Server errors
	ErrInvalidOperation = errors.New("invalid operation")
	ErrInsertCSR        = errors.New("unable to insert CSR")
	ErrInsertCert       = errors.New("unable to insert certificate")
	ErrGetCSR           = errors.New("unable to get CSR")
	ErrGetCert          = errors.New("unable to get certificate")
	ErrUpdateCSR        = errors.New("unable to update CSR")
	ErrDeleteCSR        = errors.New("unable to delete CSR")
	ErrSignCSR          = errors.New("unable to sign CSR")
	ErrRevokeCert       = errors.New("unable to revoke certificate")
	ErrResponseEncode   = errors.New("error encoding response")
)

func NewEnrollerService(csrDBStore csrstore.DB, csrFileStore csrstore.File, certsDBStore certstore.DB, certsFileStore certstore.File, secrets secrets.Secrets, homePath string) Service {
	return &enrollerService{
		csrDBStore:     csrDBStore,
		csrFileStore:   csrFileStore,
		certsDBStore:   certsDBStore,
		certsFileStore: certsFileStore,
		secrets:        secrets,
		homePath:       homePath,
	}
}

func (s *enrollerService) Health(ctx context.Context) bool {
	return true
}

func (s *enrollerService) PostCSR(ctx context.Context, data []byte) (csrmodel.CSR, error) {
	csr, err := parseCSRDataModel(data)
	if err != nil {
		return csrmodel.CSR{}, err
	}
	csr, err = s.insertCSRInDB(csr)
	if err != nil {
		return csrmodel.CSR{}, err
	}
	err = s.insertCSRFile(data, csr.Id)
	if err != nil {
		return csrmodel.CSR{}, err
	}
	return csr, nil
}

func parseCSRDataModel(data []byte) (csrmodel.CSR, error) {
	certReq, err := crypto.ParseNewCSR(data)
	if err != nil {
		return csrmodel.CSR{}, ErrInvalidCSR
	}
	csr := csrmodel.CSR{
		CountryName:            strings.Join(certReq.Subject.Country, " "),
		StateOrProvinceName:    strings.Join(certReq.Subject.Province, " "),
		LocalityName:           strings.Join(certReq.Subject.Locality, " "),
		OrganizationName:       strings.Join(certReq.Subject.Organization, " "),
		OrganizationalUnitName: strings.Join(certReq.Subject.OrganizationalUnit, " "),
		EmailAddress:           strings.Join(certReq.EmailAddresses, " "),
		CommonName:             certReq.Subject.CommonName,
		Status:                 csrmodel.PendingStatus,
	}
	return csr, nil
}

func (s *enrollerService) insertCSRInDB(csr csrmodel.CSR) (csrmodel.CSR, error) {
	id, err := s.csrDBStore.Insert(csr)
	if err != nil {
		return csrmodel.CSR{}, ErrInsertCSR
	}
	csr.Id = id
	csr.CsrFilePath = s.homePath + "/" + strconv.Itoa(id) + ".csr"
	err = s.csrDBStore.UpdateFilePath(csr)
	if err != nil {
		return csrmodel.CSR{}, ErrInsertCSR
	}
	return csr, nil
}

func (s *enrollerService) insertCSRFile(data []byte, id int) error {
	err := s.csrFileStore.Insert(id, data)
	if err != nil {
		s.csrDBStore.Delete(id)
		return ErrInsertCSR
	}
	return nil
}

func (s *enrollerService) GetPendingCSRs(ctx context.Context) csrmodel.CSRs {
	var csrs csr.CSRs
	claims := ctx.Value(jwt.JWTClaimsContextKey).(*auth.KeycloakClaims)
	admin := containsRole(claims.RealmAccess.RoleNames, "admin")
	if admin {
		csrs = s.csrDBStore.SelectAll()
	} else {
		csrs = s.csrDBStore.SelectAllByCN(claims.PreferredUsername)
	}
	return csrs
}

func (s *enrollerService) GetPendingCSRDB(ctx context.Context, id int) (csrmodel.CSR, error) {
	c, err := s.csrDBStore.SelectByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return csrmodel.CSR{}, ErrInvalidID
		}
		return csrmodel.CSR{}, ErrGetCSR
	}
	return c, nil
}

func (s *enrollerService) GetPendingCSRFile(ctx context.Context, id int) ([]byte, error) {
	data, err := s.csrFileStore.SelectByID(id)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrInvalidID
		}
		return nil, ErrGetCSR
	}
	return data, nil
}

func (s *enrollerService) PutChangeCSRStatus(ctx context.Context, csr csrmodel.CSR, id int) (csrmodel.CSR, error) {
	var err error

	prevCSR, err := s.csrDBStore.SelectByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return csrmodel.CSR{}, ErrInvalidID
		}
		return csrmodel.CSR{}, ErrGetCSR
	}

	switch status := csr.Status; status {
	case csrmodel.ApprobedStatus:
		if prevCSR.Status == csrmodel.PendingStatus {
			err = s.approbeCSR(id, csr)
			if err != nil {
				return csrmodel.CSR{}, err
			}
		} else {
			return csrmodel.CSR{}, ErrInvalidApprobeOp
		}
	case csrmodel.RevokedStatus:
		if prevCSR.Status == csrmodel.ApprobedStatus {
			_, err = s.csrDBStore.UpdateByID(id, csr)
			if err != nil {
				return csrmodel.CSR{}, ErrUpdateCSR
			}
			err = s.revokeCert(id)
			if err != nil {
				return csrmodel.CSR{}, err
			}
		} else {
			return csrmodel.CSR{}, ErrInvalidRevokeOp
		}
	case csrmodel.DeniedStatus:
		if prevCSR.Status == csrmodel.PendingStatus {
			_, err = s.csrDBStore.UpdateByID(id, csr)
			if err != nil {
				return csrmodel.CSR{}, ErrUpdateCSR
			}
		} else {
			return csrmodel.CSR{}, ErrInvalidDenyOp
		}
	default:
		return csrmodel.CSR{}, ErrInvalidOperation
	}

	return csr, nil
}

func (s *enrollerService) revokeCert(id int) error {
	revocationDate := makeOpenSSLTime(time.Now())
	err := s.certsDBStore.Revoke(id, revocationDate)
	if err != nil {
		return ErrRevokeCert
	}
	return nil

}

func (s *enrollerService) approbeCSR(id int, csr csrmodel.CSR) error {
	csrData, err := s.readCSRFromFile(id)
	if err != nil {
		return err
	}
	crt, err := s.signCSR(csrData)
	if err != nil {
		return err
	}
	err = s.insertCertInDB(id, crt)
	if err != nil {
		return err
	}
	err = s.insertCertFile(id, crt)
	if err != nil {
		return ErrInsertCert
	}
	_, err = s.csrDBStore.UpdateByID(id, csr)
	if err != nil {
		return ErrUpdateCSR
	}
	return nil

}

func (s *enrollerService) signCSR(csr *x509.CertificateRequest) (*x509.Certificate, error) {
	crtData, err := s.secrets.SignCSR(csr)
	if err != nil {
		return nil, ErrSignCSR
	}

	crt, err := x509.ParseCertificate(crtData)
	if err != nil {
		return nil, ErrSignCSR
	}
	return crt, nil
}

func (s *enrollerService) readCSRFromFile(id int) (*x509.CertificateRequest, error) {
	csrData, err := s.csrFileStore.SelectByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidID
		}
		return nil, ErrGetCSR
	}
	csr, err := crypto.ParseNewCSR(csrData)
	if err != nil {
		return nil, ErrGetCSR
	}
	return csr, nil
}

func (s *enrollerService) insertCertInDB(id int, crt *x509.Certificate) error {
	dn := makeDn(crt)
	expirationDate := makeOpenSSLTime(crt.NotAfter)
	serialHex := fmt.Sprintf("%x", crt.SerialNumber)
	certPath := s.homePath + "/" + crt.Subject.CommonName + "." + serialHex + ".crt"

	cert := certs.CRT{
		ID:             id,
		DN:             dn,
		ExpirationDate: expirationDate,
		Serial:         crt.SerialNumber,
		RevocationDate: "",
		CertPath:       certPath,
		Status:         "V",
	}
	err := s.certsDBStore.Insert(cert)
	if err != nil {
		return ErrInsertCert
	}
	return nil
}

func (s *enrollerService) insertCertFile(id int, crt *x509.Certificate) error {
	err := s.certsFileStore.Insert(id, crt.Raw)
	if err != nil {
		return ErrInsertCert
	}
	return nil
}

func (s *enrollerService) DeleteCSR(ctx context.Context, id int) error {
	csr, err := s.csrDBStore.SelectByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidID
		}
		return ErrGetCSR
	}
	if csr.Status == csrmodel.DeniedStatus || csr.Status == csrmodel.RevokedStatus {
		err = s.csrDBStore.Delete(id)
		if err != nil {
			return ErrDeleteCSR
		}
		err = s.csrFileStore.Delete(id)
		if err != nil {
			return ErrDeleteCSR
		}
		return nil
	}
	return ErrInvalidDeleteOp
}

func (s *enrollerService) GetCRT(ctx context.Context, id int) ([]byte, error) {
	data, err := s.certsFileStore.SelectByID(id)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrInvalidID
		}
		return nil, ErrGetCert
	}
	return data, nil
}

func makeDn(cert *x509.Certificate) string {
	var dn bytes.Buffer

	if len(cert.Subject.Country) > 0 && len(cert.Subject.Country[0]) > 0 {
		dn.WriteString("/C=" + cert.Subject.Country[0])
	}
	if len(cert.Subject.Province) > 0 && len(cert.Subject.Province[0]) > 0 {
		dn.WriteString("/ST=" + cert.Subject.Province[0])
	}
	if len(cert.Subject.Locality) > 0 && len(cert.Subject.Locality[0]) > 0 {
		dn.WriteString("/L=" + cert.Subject.Locality[0])
	}
	if len(cert.Subject.Organization) > 0 && len(cert.Subject.Organization[0]) > 0 {
		dn.WriteString("/O=" + cert.Subject.Organization[0])
	}
	if len(cert.Subject.OrganizationalUnit) > 0 && len(cert.Subject.OrganizationalUnit[0]) > 0 {
		dn.WriteString("/OU=" + cert.Subject.OrganizationalUnit[0])
	}
	if len(cert.Subject.CommonName) > 0 {
		dn.WriteString("/CN=" + cert.Subject.CommonName)
	}
	if len(cert.EmailAddresses) > 0 {
		dn.WriteString("/emailAddress=" + cert.EmailAddresses[0])
	}
	return dn.String()
}

func makeOpenSSLTime(t time.Time) string {
	y := (int(t.Year()) % 100)
	validDate := fmt.Sprintf("%02d%02d%02d%02d%02d%02dZ", y, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return validDate
}

func containsRole(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
