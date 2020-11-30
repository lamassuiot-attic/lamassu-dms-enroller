package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"enroller/pkg/enroller/auth"
	"enroller/pkg/enroller/configs"
	"enroller/pkg/enroller/crypto"
	certstore "enroller/pkg/enroller/models/certs/store"
	certsdb "enroller/pkg/enroller/models/certs/store/db"
	certsfile "enroller/pkg/enroller/models/certs/store/file"
	csrmodel "enroller/pkg/enroller/models/csr"
	csrstore "enroller/pkg/enroller/models/csr/store"
	csrdb "enroller/pkg/enroller/models/csr/store/db"
	csrfile "enroller/pkg/enroller/models/csr/store/file"
	"enroller/pkg/enroller/secrets"
	secretsfile "enroller/pkg/enroller/secrets/file"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/go-kit/kit/auth/jwt"
)

type serviceSetUp struct {
	csrdb    csrstore.DB
	csrfile  csrstore.File
	certdb   certstore.DB
	certfile certstore.File
	secrets  secrets.Secrets
	homePath string
}

func TestPostCSR(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	testCases := []struct {
		name string
		csr  []byte
		ret  error
	}{
		{"Correct CSR", testCSR(), nil},
		{"Incorrect CSR", []byte("This is not a CSR"), ErrInvalidCSR},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			csr, err := srv.PostCSR(ctx, []byte(tc.csr))
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
			if err == nil {
				err = stu.csrdb.Delete(csr.Id)
				if err != nil {
					t.Fatal("Could not delete CSR from DB")
				}

				err = stu.csrfile.Delete(csr.Id)
				if err != nil {
					t.Fatal("Could not delete CSR from file system")
				}
			}
		})
	}

}

func TestGetPendingCSRs(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	certReq, err := crypto.ParseNewCSR(testCSR())
	if err != nil {
		t.Fatal("Could not parse CSR")
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
	id, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}

	testCases := []struct {
		name    string
		ctx     context.Context
		numCSRs int
		cn      string
		cnRegex *regexp.Regexp
	}{
		{
			"CN = test.com manufacturing system CSRs",
			context.WithValue(ctx, jwt.JWTClaimsContextKey, &auth.KeycloakClaims{PreferredUsername: "test.com"}),
			1,
			"test.com",
			regexp.MustCompile(`^test.com$`),
		},
		{
			"Admin role CSRs",
			context.WithValue(ctx, jwt.JWTClaimsContextKey, &auth.KeycloakClaims{RealmAccess: auth.Roles{RoleNames: []string{"admin"}}}),
			1,
			"*",
			regexp.MustCompile(`^*$`),
		},
		{
			"CN = other manufacturing manufacturing system CSRs",
			context.WithValue(ctx, jwt.JWTClaimsContextKey, &auth.KeycloakClaims{PreferredUsername: "other"}),
			0,
			"other",
			regexp.MustCompile(`^other$`),
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			csrs := srv.GetPendingCSRs(tc.ctx)
			if tc.numCSRs != len(csrs.CSRs) {
				t.Errorf("Got number of CSRs is %d; want %d", len(csrs.CSRs), tc.numCSRs)
			}
			if len(csrs.CSRs) > 0 {
				if !tc.cnRegex.MatchString(csrs.CSRs[0].CommonName) {
					t.Errorf("Evaluated CN regex pattern does not return any result. Got %s; want %s", tc.cn, csrs.CSRs[0].CommonName)
				}
			}
		})
	}

	err = stu.csrdb.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from DB")
	}
}

func TestGetPendingCSRDB(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	certReq, err := crypto.ParseNewCSR(testCSR())
	if err != nil {
		t.Fatal("Could not parse CSR")
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
	id, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}

	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrInvalidID},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetPendingCSRDB(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = stu.csrdb.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from DB")
	}
}

func TestGetPendingCSRFile(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	certReq := testCSR()
	id := 1
	err := stu.csrfile.Insert(id, certReq)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}

	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrInvalidID},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetPendingCSRFile(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}

	err = stu.csrfile.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from file system")
	}
}

func TestPutChangeCSRStatus(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	csrRaw := testCSR()
	certReq, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
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
	id, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}

	err = stu.csrfile.Insert(id, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}

	testCases := []struct {
		name   string
		status string
		id     int
		csr    csrmodel.CSR
		ret    error
	}{
		{"Revoke NEW Status CSR", csrmodel.RevokedStatus, id, csr, ErrInvalidRevokeOp},
		{"Approbe NEW Status CSR ID does not exist", csrmodel.ApprobedStatus, id + 1000, csr, ErrInvalidID},
		{"Approbe NEW Status CSR ID exists", csrmodel.ApprobedStatus, id, csr, nil},
		{"Deny APPROBED Status CSR", csrmodel.DeniedStatus, id, csr, ErrInvalidDenyOp},
		{"Revoke APPROBED Status CSR", csrmodel.RevokedStatus, id, csr, nil},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			tc.csr.Status = tc.status
			retcsr, err := srv.PutChangeCSRStatus(ctx, tc.csr, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
			if err == nil {
				if tc.csr.Status != retcsr.Status {
					t.Errorf("Got result is %s; want %s", retcsr.Status, tc.csr.Status)
				}
			}

		})
	}

	err = stu.csrdb.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from DB")
	}

	err = stu.csrfile.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from file system")
	}

	err = stu.certdb.Delete(id)
	if err != nil {
		t.Fatal("Could not delete certificate from DB")
	}

	err = stu.certfile.Delete(id)
	if err != nil {
		t.Fatal("Could not delete certificate from file system")
	}
}

func TestGetCRT(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	csrRaw := testCSR()
	certReq, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
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
	id, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in DB")
	}

	crtData, err := stu.secrets.SignCSR(certReq)
	if err != nil {
		t.Fatal("Could not sign CSR")
	}

	err = stu.certfile.Insert(id, crtData)
	if err != nil {
		t.Fatal("Could not insert certificate in file system")
	}

	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrInvalidID},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetCRT(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}

	err = stu.csrdb.Delete(id)
	if err != nil {
		t.Fatal("Could not delete CSR from DB")
	}

	err = stu.certfile.Delete(id)
	if err != nil {
		t.Fatal("Could not delete certificate from file system")
	}
}

func TestDelete(t *testing.T) {
	stu := setup()
	srv := NewEnrollerService(stu.csrdb, stu.csrfile, stu.certdb, stu.certfile, stu.secrets, stu.homePath)
	ctx := context.Background()

	csrRaw := testCSR()
	certReq, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
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
	newID, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in DB")
	}

	err = stu.csrfile.Insert(newID, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}

	approbeID, err := stu.csrdb.Insert(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in DB")
	}

	csr.Status = csrmodel.ApprobedStatus
	_, err = stu.csrdb.UpdateByID(approbeID, csr)
	if err != nil {
		t.Fatal("Could not update CSR status in DB")
	}

	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"Delete CSR status is not NEW", approbeID, ErrInvalidDeleteOp},
		{"Delete CSR ID does not exist", newID + 1000, ErrInvalidID},
		{"Delete CSR ID exists", newID, nil},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			err := srv.DeleteCSR(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}

	err = stu.csrdb.Delete(approbeID)
	if err != nil {
		t.Fatal("Could not delete CSR from DB")
	}
}

func setup() *serviceSetUp {
	err, cfg := configs.NewConfig("enrollertest")
	if err != nil {
		panic(err)
	}
	connStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	csrdb, err := setupCSRDB(connStr)
	if err != nil {
		panic(err)
	}
	certdb, err := setupCertDB(connStr)
	if err != nil {
		panic(err)
	}
	csrfile := setupCSRFile(cfg.HomePath)
	certfile := setupCertFile(cfg.HomePath)
	secrets := setupSecrets(cfg.CACertFile, cfg.CAKeyFile, cfg.OCSPServer, certdb)
	return &serviceSetUp{csrdb, csrfile, certdb, certfile, secrets, cfg.HomePath}
}

func setupCSRDB(connStr string) (csrstore.DB, error) {
	db, err := csrdb.NewDB("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupCertDB(connStr string) (certstore.DB, error) {
	db, err := certsdb.NewDB("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupCSRFile(path string) csrstore.File {
	return csrfile.NewFile(path)
}

func setupCertFile(path string) certstore.File {
	return certsfile.NewFile(path)
}

func setupSecrets(CACertFile string, CAKeyFile string, OCSPServer string, certsdb certstore.DB) secrets.Secrets {
	return secretsfile.NewFile(CACertFile, CAKeyFile, OCSPServer, certsdb)
}

func testCSR() []byte {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 1024)

	subj := pkix.Name{
		CommonName:   "test.com",
		Country:      []string{"ES"},
		Province:     []string{"Gipuzkoa"},
		Locality:     []string{"Arrasate"},
		Organization: []string{"Test"},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		panic(err)
	}
	csr := new(bytes.Buffer)
	pem.Encode(csr, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csr.Bytes()
}
