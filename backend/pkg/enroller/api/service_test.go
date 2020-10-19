package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"enroller/pkg/enroller/configs"
	"enroller/pkg/enroller/crypto"
	"enroller/pkg/enroller/models/db"
	"enroller/pkg/enroller/models/files"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestPostCSR(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	testCases := []struct {
		name string
		csr  []byte
		ret  error
	}{
		{"Correct CSR", testCSR(), nil},
		{"Incorrect CSR", []byte("This is not a CSR"), ErrIncorrectContent},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.PostCSR(ctx, []byte(tc.csr))
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err := teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestGetPendingCSRs(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csr, err := crypto.ParseNewCSR(testCSR())
	if err != nil {
		t.Fatal("Could not parse CSR")
	}
	_, err = db.InsertCSR(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}
	csrs := srv.GetPendingCSRs(ctx)
	fmt.Println(csrs)
	if len(csrs.CSRs) <= 0 {
		t.Errorf("Not CSRs returned from API")
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestGetPendingCSRDB(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csr, err := crypto.ParseNewCSR(testCSR())
	if err != nil {
		t.Fatal("Could not parse CSR")
	}
	id, err := db.InsertCSR(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}
	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrIncorrectContent},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetPendingCSRDB(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestGetPendingCSRFile(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csrRaw := testCSR()
	id := 1
	err := file.InsertFileCSR(id, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}
	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrIncorrectContent},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetPendingCSRFile(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestPutChangeCSRStatus(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csrRaw := testCSR()
	csr, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
	}
	id, err := db.InsertCSR(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}
	err = file.InsertFileCSR(id, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}
	testCases := []struct {
		name   string
		status string
		id     int
		csr    crypto.CSR
		ret    error
	}{
		{"CSR ID exists", "APPROBED", id, csr, nil},
		{"CSR ID does not exist", "APPROBED", id + 1000, csr, ErrBadRouting},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			tc.csr.Status = tc.status
			err := srv.PutChangeCSRStatus(ctx, tc.csr, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestDeleteCSR(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csrRaw := testCSR()
	csr, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
	}
	id, err := db.InsertCSR(csr)
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}
	err = file.InsertFileCSR(id, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}
	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrIncorrectContent},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			err := srv.DeleteCSR(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func TestGetCRT(t *testing.T) {
	db, file := setup()
	srv := NewEnrollerService(db, file)
	ctx := context.Background()
	csrRaw := testCSR()
	csr, err := crypto.ParseNewCSR(csrRaw)
	if err != nil {
		t.Fatal("Could not parse CSR")
	}
	id, err := db.InsertCSR(csr)
	csr.Id = id
	if err != nil {
		t.Fatal("Could not insert CSR in database")
	}
	err = file.InsertFileCSR(id, csrRaw)
	if err != nil {
		t.Fatal("Could not insert CSR in file system")
	}
	serial, err := file.Serial()
	if err != nil {
		t.Fatal("Could not obtain CA DB serial number")
	}
	caCertData, err := file.LoadCACert()
	if err != nil {
		t.Fatal("Could not obtain CA certificate")
	}
	caKeyData, err := file.LoadCAKey()
	if err != nil {
		t.Fatal("Could not obtain CA key")
	}
	crtData, err := csr.SignCSR(csrRaw, serial, caCertData, caKeyData)
	if err != nil {
		t.Fatal("Could not sign CSR")
	}
	err = file.InsertFileCert(csr.Id, crtData)
	if err != nil {
		t.Fatal("Could not insert certificate in file system")
	}
	testCases := []struct {
		name string
		id   int
		ret  error
	}{
		{"CSR ID exists", id, nil},
		{"CSR ID does not exist", id + 1000, ErrBadRouting},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			_, err := srv.GetCRT(ctx, tc.id)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}
	err = teardown(db, file)
	if err != nil {
		t.Fatal("Could not perform tests teardown")
	}
}

func setup() (*db.DB, *files.File) {
	err, cfg := configs.NewConfig()
	if err != nil {
		panic(err)
	}
	db, err := setupDB(cfg)
	if err != nil {
		panic(err)
	}
	file, err := setupFile(cfg)
	if err != nil {
		panic(err)
	}
	err = testCA(cfg)
	if err != nil {
		panic(err)
	}
	return db, file
}

func teardown(db *db.DB, file *files.File) error {
	err := db.TruncateTable()
	if err != nil {
		return err
	}
	err = db.CloseDB()
	if err != nil {
		return err
	}
	err = file.EmptyDirectory()
	if err != nil {
		return err
	}
	return nil
}

func setupDB(cfg configs.Config) (*db.DB, error) {
	connStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	db, err := db.NewDB("postgres", connStr)
	err = db.TruncateTable()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func setupFile(cfg configs.Config) (*files.File, error) {
	file := files.NewFile(cfg.HomePath, cfg.CAPath, cfg.CACertFile, cfg.CAKeyFile)
	err := file.EmptyDirectory()
	if err != nil {
		return nil, err
	}
	return file, nil
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

func testCA(cfg configs.Config) error {
	keyBytes, _ := rsa.GenerateKey(rand.Reader, 1024)
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			CommonName:   "test.com",
			Country:      []string{"ES"},
			Province:     []string{"Gipuzkoa"},
			Locality:     []string{"Arrasate"},
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &keyBytes.PublicKey, keyBytes)
	if err != nil {
		return err
	}
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})
	caCertPath := cfg.CAPath + "/" + cfg.CACertFile
	err = createTestFile(caCertPath, caPEM.Bytes())
	if err != nil {
		return err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(keyBytes)})
	caKeyPath := cfg.CAPath + "/" + cfg.CAKeyFile
	err = createTestFile(caKeyPath, caPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func createTestFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		os.Remove(path)
		return err
	}
	return nil
}
