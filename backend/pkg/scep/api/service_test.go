package api

import (
	"context"
	"enroller/pkg/scep/configs"
	"enroller/pkg/scep/crypto"
	"enroller/pkg/scep/models/db"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

type serviceSetUp struct {
	scepDB db.DBSCEPStore
}

func TestGetSCEPCRTs(t *testing.T) {
	stu := setup()
	srv := NewSCEPService(stu.scepDB)
	ctx := context.Background()

	crt := testCRT()
	err := stu.scepDB.InsertCRT(crt)
	if err != nil {
		t.Fatal("Could not insert certificate in DB")
	}

	crts, err := srv.GetSCEPCRTs(ctx)
	if err != nil {
		t.Errorf("SCEP API returned error: %s", err)
	}
	if len(crts.CRTs) <= 0 {
		t.Errorf("Not certificates returned from SCEP API")
	}

	err = stu.scepDB.Delete(crt.DN, crt.Serial)
	if err != nil {
		t.Fatal("Could not delete certificate from DB")
	}
}

func TestRevokeSCEPCRT(t *testing.T) {
	stu := setup()
	srv := NewSCEPService(stu.scepDB)
	ctx := context.Background()

	crt := testCRT()
	err := stu.scepDB.InsertCRT(crt)
	if err != nil {
		t.Fatal("Could not insert certificate in DB")
	}

	testCases := []struct {
		name   string
		dn     string
		serial string
		ret    error
	}{

		{
			"Revoke certificate DN does not exist",
			"doesNotExist",
			crt.Serial,
			ErrInvalidDNOrSerial,
		},
		{
			"Revoke certificate Serial does not exist",
			crt.DN,
			"100000",
			ErrInvalidDNOrSerial,
		},
		{
			"Revoke certificate exists",
			crt.DN,
			crt.Serial,
			nil,
		},
		{
			"Revoke revoked certificate",
			crt.DN,
			crt.Serial,
			ErrInvalidRevokeOp,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Testing %s", tc.name), func(t *testing.T) {
			err := srv.RevokeSCEPCRT(ctx, tc.dn, tc.serial)
			if tc.ret != err {
				t.Errorf("Got result is %s; want %s", err, tc.ret)
			}
		})
	}

	err = stu.scepDB.Delete(crt.DN, crt.Serial)
	if err != nil {
		t.Fatal("Could not delete certificate from DB")
	}
}

func setup() *serviceSetUp {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}
	err, cfg := configs.NewConfig("sceptest")
	if err != nil {
		panic(err)
	}
	connStr := "dbname=" + cfg.PostgresDB + " user=" + cfg.PostgresUser + " password=" + cfg.PostgresPassword + " host=" + cfg.PostgresHostname + " port=" + cfg.PostgresPort + " sslmode=disable"
	scepDB, err := setupSCEPDB(connStr, logger)
	if err != nil {
		panic(err)
	}
	return &serviceSetUp{scepDB}
}

func setupSCEPDB(connStr string, logger log.Logger) (db.DBSCEPStore, error) {
	db, err := db.NewDB("postgres", connStr, logger)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func testCRT() crypto.CRT {
	crt := crypto.CRT{
		Status:         "V",
		ExpirationDate: makeOpenSSLTime(time.Now()),
		RevocationDate: "",
		Serial:         fmt.Sprintf("%x", big.NewInt(1)),
		DN:             "/CN=test",
		CRTPath:        "/tmp/test.crt",
		Key:            "RSA",
		KeySize:        2048,
	}
	return crt
}

func makeOpenSSLTime(t time.Time) string {
	y := (int(t.Year()) % 100)
	validDate := fmt.Sprintf("%02d%02d%02d%02d%02d%02dZ", y, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return validDate
}
