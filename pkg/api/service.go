package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"

	"github.com/go-kit/kit/log"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/client"

	"github.com/lamassuiot/dms-enroller/pkg/models/dms"
	dmsstore "github.com/lamassuiot/dms-enroller/pkg/models/dms/store"
	"github.com/lamassuiot/dms-enroller/pkg/utils"
)

type Service interface {
	Health(ctx context.Context) bool
	CreateDMS(ctx context.Context, csrBase64Encoded string, dmsName string) (dms.DMS, error)
	CreateDMSForm(ctx context.Context, dmsForm dms.DmsCreationForm) (string, dms.DMS, error)
	UpdateDMSStatus(ctx context.Context, dms dms.DMS, id int) (dms.DMS, error)
	DeleteDMS(ctx context.Context, id int) error
	GetDMSs(ctx context.Context) ([]dms.DMS, error)
	GetDMSCertificate(ctx context.Context, id int) (*x509.Certificate, error)
}

type enrollerService struct {
	mtx             sync.RWMutex
	dmsDBStore      dmsstore.DB
	lamassuCaClient lamassucaclient.LamassuCaClient
	logger          log.Logger
}

var (
	// Client errors
	ErrInvalidCSRBase64Format = errors.New("unable to decode base64 encoded CSR") //400
	ErrInvalidCSR             = errors.New("unable to parse CSR, is invalid")     //400
	ErrInvalidID              = errors.New("invalid CSR ID, does not exist")      //404
	ErrInvalidIDFormat        = errors.New("invalid ID format")
	ErrInvalidApprobeOp       = errors.New("invalid operation, only pending status CSRs can be approved")          //400
	ErrInvalidRevokeOp        = errors.New("invalid operation, only approved status CSRs can be revoked")          //400
	ErrInvalidDenyOp          = errors.New("invalid operation, only pending status CSRs can be denied")            //400
	ErrInvalidDeleteOp        = errors.New("invalid operation, only denied or revoked status CSRs can be deleted") //400
	ErrIncorrectType          = errors.New("unsupported media type")                                               //415
	ErrEmptyDMSName           = errors.New("empty DMS name")                                                       //415
	ErrEmptyBody              = errors.New("empty body")

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

func NewEnrollerService(dmsDbStore dmsstore.DB, lamassuCa *lamassucaclient.LamassuCaClient, logger log.Logger) Service {
	return &enrollerService{
		dmsDBStore:      dmsDbStore,
		lamassuCaClient: *lamassuCa,
		logger:          logger,
	}
}

func (s *enrollerService) Health(ctx context.Context) bool {
	return true
}

func (s *enrollerService) CreateDMS(ctx context.Context, csrBase64Encoded string, dmsName string) (dms.DMS, error) {

	//csrBase64Encoded
	decodedCsr, err := utils.DecodeB64(csrBase64Encoded)
	if err != nil {
		return dms.DMS{}, err
	}

	p, _ := pem.Decode([]byte(decodedCsr))
	if err != nil {
		return dms.DMS{}, err
	}
	if err != nil {
		return dms.DMS{}, err
	}

	csr, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return dms.DMS{}, err
	}

	keyType, keyBits := getPublicKeyInfo(csr)

	d := dms.DMS{
		Name:      dmsName,
		CsrBase64: csrBase64Encoded,
		Status:    dms.PendingStatus,
		KeyType:   keyType,
		KeyBits:   keyBits,
	}

	dmsId, err := s.dmsDBStore.Insert(ctx, d)

	if err != nil {
		return dms.DMS{}, err
	}

	return s.dmsDBStore.SelectByID(ctx, dmsId)
}

func (s *enrollerService) CreateDMSForm(ctx context.Context, dmsForm dms.DmsCreationForm) (string, dms.DMS, error) {
	subj := pkix.Name{
		CommonName:         dmsForm.CommonName,
		Country:            []string{dmsForm.CountryName},
		Province:           []string{dmsForm.StateOrProvinceName},
		Locality:           []string{dmsForm.LocalityName},
		Organization:       []string{dmsForm.OrganizationName},
		OrganizationalUnit: []string{dmsForm.OrganizationalUnitName},
	}

	if dmsForm.KeyType == "rsa" {
		privKey, _ := rsa.GenerateKey(rand.Reader, dmsForm.KeyBits)
		csrBytes, err := generateCSR(ctx, dmsForm.KeyType, dmsForm.KeyBits, privKey, subj)
		if err != nil {
			return "", dms.DMS{}, err
		}

		csrEncoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

		privkey_bytes := x509.MarshalPKCS1PrivateKey(privKey)
		privkey_pem := string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: privkey_bytes,
			},
		))

		csr, err := s.CreateDMS(ctx, utils.EcodeB64(string(csrEncoded)), dmsForm.Name)
		if err != nil {
			return "", dms.DMS{}, err
		} else {
			return privkey_pem, csr, nil
		}
	} else if dmsForm.KeyType == "ec" {
		var priv *ecdsa.PrivateKey
		var err error
		switch dmsForm.KeyBits {
		case 224:
			priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		case 256:
			priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		case 384:
			priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		case 521:
			priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		default:
			err = errors.New("Unsupported key length")
		}
		if err != nil {
			return "", dms.DMS{}, err
		}
		privkey_bytesm, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return "", dms.DMS{}, err
		}
		privkey_pem := string(pem.EncodeToMemory(
			&pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: privkey_bytesm,
			},
		))
		csrBytes, err := generateCSR(ctx, dmsForm.KeyType, dmsForm.KeyBits, priv, subj)
		if err != nil {
			return "", dms.DMS{}, err
		}
		csrEncoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
		csr, err := s.CreateDMS(ctx, string(csrEncoded), dmsForm.Name)
		if err != nil {
			return "", dms.DMS{}, err
		} else {
			return privkey_pem, csr, nil
		}
	} else {
		return "", dms.DMS{}, errors.New("Invalid key format")
	}
}

func generateCSR(ctx context.Context, keyType string, keyBits int, priv interface{}, subj pkix.Name) ([]byte, error) {
	var signingAlgorithm x509.SignatureAlgorithm
	if keyType == "ec" {
		signingAlgorithm = x509.ECDSAWithSHA512
	} else {
		signingAlgorithm = x509.SHA512WithRSA
	}
	rawSubj := subj.ToRDNSequence()
	/*rawSubj = append(rawSubj, []pkix.AttributeTypeAndValue{
		{Type: oidEmailAddress, Value: emailAddress},
	})*/

	asn1Subj, _ := asn1.Marshal(rawSubj)
	template := x509.CertificateRequest{
		RawSubject: asn1Subj,
		//EmailAddresses:     []string{emailAddress},
		SignatureAlgorithm: signingAlgorithm,
	}
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, priv)
	return csrBytes, err
}

func (s *enrollerService) UpdateDMSStatus(ctx context.Context, d dms.DMS, id int) (dms.DMS, error) {
	var err error

	prevDms, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return dms.DMS{}, ErrInvalidID
		}
		return dms.DMS{}, ErrGetCSR
	}

	switch status := d.Status; status {
	case dms.ApprovedStatus:
		if prevDms.Status == dms.PendingStatus {
			b, err := utils.DecodeB64(prevDms.CsrBase64)
			if err != nil {
				return dms.DMS{}, err
			}
			csr, err := x509.ParseCertificateRequest([]byte(b))
			if err != nil {
				return dms.DMS{}, err
			}
			crt, err := s.ApprobeCSR(ctx, id, csr)
			if err != nil {
				return dms.DMS{}, err
			}
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.ApprovedStatus, InsertNth(ToHexInt(crt.SerialNumber), 2), "")
			if err != nil {
				return dms.DMS{}, ErrUpdateCSR
			}

			if err != nil {
				return dms.DMS{}, err
			}
		} else {
			return dms.DMS{}, ErrInvalidApprobeOp
		}
	case dms.RevokedStatus:
		if prevDms.Status == dms.ApprovedStatus {
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.RevokedStatus, prevDms.SerialNumber, "")
			if err != nil {
				return dms.DMS{}, ErrUpdateCSR
			}
			err = s.RevokeCert(ctx, prevDms.SerialNumber)
			if err != nil {
				return dms.DMS{}, err
			}
		} else {
			return dms.DMS{}, ErrInvalidRevokeOp
		}
	case dms.DeniedStatus:
		if prevDms.Status == dms.PendingStatus {
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.DeniedStatus, "", "")
			if err != nil {
				return dms.DMS{}, ErrUpdateCSR
			}
		} else {
			return dms.DMS{}, ErrInvalidDenyOp
		}
	default:
		return dms.DMS{}, ErrInvalidOperation
	}

	return d, nil
}

func (s *enrollerService) RevokeCert(ctx context.Context, serialToRevoke string) error {
	// revocar llamando a lamassu CA
	err := s.lamassuCaClient.RevokeCert(ctx, "Lamassu-DMS-Enroller", serialToRevoke, "dmsenroller")
	return err
}

func (s *enrollerService) ApprobeCSR(ctx context.Context, id int, csr *x509.CertificateRequest) (*x509.Certificate, error) {

	crt, err := s.lamassuCaClient.SignCertificateRequest(ctx, "Lamassu-DMS-Enroller", csr, "dmsenroller")
	if err != nil {
		return &x509.Certificate{}, ErrSignCSR
	}

	return crt, nil
}

func (s *enrollerService) DeleteDMS(ctx context.Context, id int) error {
	d, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidID
		}
		return ErrGetCSR
	}
	if d.Status == dms.DeniedStatus || d.Status == dms.RevokedStatus {
		err = s.dmsDBStore.Delete(ctx, id)
		if err != nil {
			return ErrDeleteCSR
		}
		err = s.dmsDBStore.Delete(ctx, id)
		if err != nil {
			return ErrDeleteCSR
		}
		return nil
	}
	return ErrInvalidDeleteOp
}
func (s *enrollerService) GetDMSs(ctx context.Context) ([]dms.DMS, error) {
	d, err := s.dmsDBStore.SelectAll(ctx)
	if err != nil {
		return []dms.DMS{}, nil
	}
	return d, nil
}

func (s *enrollerService) GetDMSCertificate(ctx context.Context, id int) (*x509.Certificate, error) {
	d, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrInvalidID
		}
		return nil, ErrGetCert
	}

	lamassuCert, err := s.lamassuCaClient.GetCert(ctx, "Lamassu-DMS-Enroller", d.SerialNumber, "dmsenroller")

	if err != nil {
		return &x509.Certificate{}, ErrSignCSR
	}

	decodedCert, err := base64.StdEncoding.DecodeString(lamassuCert.CertContent.CerificateBase64)

	p, _ := pem.Decode(decodedCert)
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return cert, nil
}

func containsRole(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func InsertNth(s string, n int) string {
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

func ToHexInt(n *big.Int) string {
	return fmt.Sprintf("%x", n) // or %X or upper case
}

func getPublicKeyInfo(cert *x509.CertificateRequest) (string, int) {
	key := cert.PublicKeyAlgorithm.String()
	var keyBits int
	switch key {
	case "RSA":
		keyBits = cert.PublicKey.(*rsa.PublicKey).N.BitLen()
		return "RSA", keyBits
	case "ECDSA":
		keyBits = cert.PublicKey.(*ecdsa.PublicKey).Params().BitSize
		return "ECDSA", keyBits
	}

	return "UNKOWN_KEY_TYPE", -1
}
