package service

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"sync"

	"github.com/go-kit/kit/log"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/pkg/client"

	"github.com/lamassuiot/dms-enroller/pkg/server/models/dms"
	dmsstore "github.com/lamassuiot/dms-enroller/pkg/server/models/dms/store"
	"github.com/lamassuiot/dms-enroller/pkg/server/utils"
)

type Service interface {
	Health(ctx context.Context) bool
	CreateDMS(ctx context.Context, csrBase64Encoded string, dmsName string) (dms.DMS, error)
	CreateDMSForm(ctx context.Context, subject dms.Subject, PrivateKeyMetadata dms.PrivateKeyMetadata, url string, dmsName string) (string, dms.DMS, error)
	UpdateDMSStatus(ctx context.Context, status string, id int) (dms.DMS, error)
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

	csr, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return dms.DMS{}, err
	}

	keyType, keyBits := getPublicKeyInfo(csr)

	d := dms.DMS{
		Name:      dmsName,
		CsrBase64: csrBase64Encoded,
		Status:    dms.PendingStatus,
		KeyMetadata: dms.PrivateKeyMetadata{
			KeyType: keyType,
			KeyBits: keyBits,
		},
	}

	dmsId, err := s.dmsDBStore.Insert(ctx, d)

	if err != nil {
		return dms.DMS{}, err
	}

	return s.dmsDBStore.SelectByID(ctx, dmsId)
}

func (s *enrollerService) CreateDMSForm(ctx context.Context, subject dms.Subject, PrivateKeyMetadata dms.PrivateKeyMetadata, url string, dmsName string) (string, dms.DMS, error) {
	subj := pkix.Name{
		CommonName:         subject.CN,
		Country:            []string{subject.C},
		Province:           []string{subject.ST},
		Locality:           []string{subject.L},
		Organization:       []string{subject.O},
		OrganizationalUnit: []string{subject.OU},
	}

	if PrivateKeyMetadata.KeyType == "rsa" {
		privKey, _ := rsa.GenerateKey(rand.Reader, PrivateKeyMetadata.KeyBits)
		csrBytes, err := generateCSR(ctx, PrivateKeyMetadata.KeyType, PrivateKeyMetadata.KeyBits, privKey, subj)
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
		privkey_pem = utils.EcodeB64(privkey_pem)
		csr, err := s.CreateDMS(ctx, utils.EcodeB64(string(csrEncoded)), dmsName)
		if err != nil {
			return "", dms.DMS{}, err
		} else {
			return privkey_pem, csr, nil
		}
	} else if PrivateKeyMetadata.KeyType == "ec" {
		var priv *ecdsa.PrivateKey
		var err error
		switch PrivateKeyMetadata.KeyBits {
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
		privkey_pem = utils.EcodeB64(privkey_pem)
		csrBytes, err := generateCSR(ctx, PrivateKeyMetadata.KeyType, PrivateKeyMetadata.KeyBits, priv, subj)
		if err != nil {
			return "", dms.DMS{}, err
		}
		csrEncoded := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
		csr, err := s.CreateDMS(ctx, string(csrEncoded), dmsName)
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

func (s *enrollerService) UpdateDMSStatus(ctx context.Context, DMSstatus string, id int) (dms.DMS, error) {
	var err error
	var d dms.DMS
	prevDms, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		return dms.DMS{}, err
	}

	switch status := DMSstatus; status {
	case dms.ApprovedStatus:
		if prevDms.Status == dms.PendingStatus {
			b, err := utils.DecodeB64(prevDms.CsrBase64)
			if err != nil {
				return dms.DMS{}, err
			}
			csrBytes, _ := pem.Decode([]byte(b))
			csr, err := x509.ParseCertificateRequest(csrBytes.Bytes)
			if err != nil {
				return dms.DMS{}, err
			}
			crt, err := s.ApprobeCSR(ctx, id, csr)
			if err != nil {
				return dms.DMS{}, err
			}
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.ApprovedStatus, InsertNth(ToHexInt(crt.SerialNumber), 2), "")
			if err != nil {
				return dms.DMS{}, err
			}
			var cb []byte
			cb = append(cb, crt.Raw...)
			certificate := pem.Block{Type: "CERTIFICATE", Bytes: cb}
			cert := pem.EncodeToMemory(&certificate)

			d.CerificateBase64 = utils.EcodeB64(string(cert))

		} else {
			return dms.DMS{}, err
		}
	case dms.RevokedStatus:
		if prevDms.Status == dms.ApprovedStatus {
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.RevokedStatus, prevDms.SerialNumber, "")
			if err != nil {
				return dms.DMS{}, err
			}
			err = s.RevokeCert(ctx, prevDms.SerialNumber)
			if err != nil {
				return dms.DMS{}, err
			}
		} else {
			return dms.DMS{}, err
		}
	case dms.DeniedStatus:
		if prevDms.Status == dms.PendingStatus {
			d, err = s.dmsDBStore.UpdateByID(ctx, id, dms.DeniedStatus, "", "")
			if err != nil {
				return dms.DMS{}, err
			}
		} else {
			return dms.DMS{}, err
		}
	default:
		return dms.DMS{}, err
	}

	return d, nil
}

func (s *enrollerService) RevokeCert(ctx context.Context, serialToRevoke string) error {
	// revocar llamando a lamassu CA
	err := s.lamassuCaClient.RevokeCert(ctx, "Lamassu-DMS-Enroller", serialToRevoke, "dmsenroller")
	return err
}

func (s *enrollerService) ApprobeCSR(ctx context.Context, id int, csr *x509.CertificateRequest) (*x509.Certificate, error) {

	crt, err := s.lamassuCaClient.SignCertificateRequest(ctx, "Lamassu-DMS-Enroller", csr, "dmsenroller", true)
	if err != nil {
		return &x509.Certificate{}, err
	}

	return crt, nil
}

func (s *enrollerService) DeleteDMS(ctx context.Context, id int) error {
	d, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		return err
	}
	if d.Status == dms.DeniedStatus || d.Status == dms.RevokedStatus {
		err = s.dmsDBStore.Delete(ctx, id)
		if err != nil {
			return err
		}
	}
	return err
}
func (s *enrollerService) GetDMSs(ctx context.Context) ([]dms.DMS, error) {
	d, err := s.dmsDBStore.SelectAll(ctx)
	if err != nil {
		return []dms.DMS{}, err
	}
	var dmsList []dms.DMS
	for _, item := range d {
		lamassuCert, _ := s.lamassuCaClient.GetCert(ctx, "Lamassu-DMS-Enroller", item.SerialNumber, "dmsenroller")
		item.Subject = dms.Subject{
			C:  lamassuCert.Subject.C,
			ST: lamassuCert.Subject.ST,
			L:  lamassuCert.Subject.L,
			O:  lamassuCert.Subject.O,
			OU: lamassuCert.Subject.OU,
			CN: lamassuCert.Subject.CN,
		}
		item.CerificateBase64 = lamassuCert.CertContent.CerificateBase64
		dmsList = append(dmsList, item)
	}
	return dmsList, nil
}

func (s *enrollerService) GetDMSCertificate(ctx context.Context, id int) (*x509.Certificate, error) {
	d, err := s.dmsDBStore.SelectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	lamassuCert, err := s.lamassuCaClient.GetCert(ctx, "Lamassu-DMS-Enroller", d.SerialNumber, "dmsenroller")

	if err != nil {
		return &x509.Certificate{}, err
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
