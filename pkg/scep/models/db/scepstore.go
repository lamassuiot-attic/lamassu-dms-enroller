package db

import (
	"database/sql"
	"enroller/pkg/scep/crypto"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	_ "github.com/lib/pq"
)

func NewDB(driverName string, dataSourceName string, logger log.Logger) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not open connection with signed certificates database")
		return nil, err
	}
	err = checkDBAlive(db)
	for err != nil {
		level.Warn(logger).Log("msg", "Trying to connect to CSRs DB")
		err = checkDBAlive(db)
	}

	return &DB{db, logger}, nil
}

type DB struct {
	*sql.DB
	logger log.Logger
}

func checkDBAlive(db *sql.DB) error {
	sqlStatement := `
	SELECT WHERE 1=0`
	_, err := db.Query(sqlStatement)
	return err
}

func (db *DB) InsertCRT(crt crypto.CRT) error {
	sqlStatement := `

	INSERT INTO ca_store(status, expirationDate, revocationDate, serial, dn, certPath, key, keySize)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING serial;
	`
	serialHex := fmt.Sprintf("%x", crt.Serial)
	var serial string

	err := db.QueryRow(sqlStatement, crt.Status, crt.ExpirationDate, crt.RevocationDate, serialHex, crt.DN, crt.CRTPath, crt.Key, crt.KeySize).Scan(&serial)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not insert certificate with serial "+crt.Serial+" in database")
		return err
	}
	level.Info(db.logger).Log("msg", "Certificate with serial "+serial+" inserted in database")

	return nil
}

func (db *DB) SelectCRT(dn string, serial string) (crypto.CRT, error) {
	sqlStatement := `
	SELECT *
	FROM ca_store
	WHERE dn = $1 AND serial = $2;
	`
	serialHex := fmt.Sprintf("%x", serial)

	row := db.QueryRow(sqlStatement, dn, serialHex)
	var crt crypto.CRT
	err := row.Scan(&crt.Status, &crt.ExpirationDate, &crt.RevocationDate, &crt.Serial, &crt.DN, &crt.CRTPath, &crt.Key, &crt.KeySize)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain certificate with DN "+dn+" and serial "+serial+" from database")
		return crypto.CRT{}, err
	}
	level.Info(db.logger).Log("msg", "Certificate with DN "+dn+" and serial "+serial+" read from database")
	return crt, nil
}

func (db *DB) GetCRTs() (crypto.CRTs, error) {
	sqlStatement := `
	SELECT *
	FROM ca_store;
	`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain certificates from database or the database is empty")
		return crypto.CRTs{CRTs: []crypto.CRT{}}, err
	}

	defer rows.Close()
	crts := make([]crypto.CRT, 0)

	for rows.Next() {
		var crt crypto.CRT
		err := rows.Scan(&crt.Status, &crt.ExpirationDate, &crt.RevocationDate, &crt.Serial, &crt.DN, &crt.CRTPath, &crt.Key, &crt.KeySize)
		if err != nil {
			level.Error(db.logger).Log("err", err, "msg", "Unable to read database certificate row")
			return crypto.CRTs{CRTs: []crypto.CRT{}}, err
		}
		level.Info(db.logger).Log("msg", "Certificate with serial "+crt.Serial+" read from database")
		crts = append(crts, crt)
	}

	if err = rows.Err(); err != nil {
		level.Error(db.logger).Log("err", err)
		return crypto.CRTs{CRTs: []crypto.CRT{}}, err
	}
	level.Info(db.logger).Log("msg", strconv.Itoa(len(crts))+" CSRs read from database")
	return crypto.CRTs{CRTs: crts}, nil
}

func (db *DB) RevokeCRT(dn string, serial string) error {
	serialHex := fmt.Sprintf("%x", serial)

	sqlStatement := `
	UPDATE ca_store
	SET status = 'R', revocationDate = $1
	WHERE dn = $2 AND serial = $3;
	`
	res, err := db.Exec(sqlStatement, makeOpenSSLTime(time.Now()), dn, serialHex)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not revoke certificate with DN "+dn+" and serial "+serial+" in database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not revoke certificate with DN "+dn+" and serial "+serial+" in database")
		return err
	}

	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}

	return nil
}

func (db *DB) Delete(dn string, serial string) error {
	sqlStatement := `
	DELETE FROM ca_store
	WHERE dn = $1 AND serial = $2;
	`
	serialHex := fmt.Sprintf("%x", serial)
	res, err := db.Exec(sqlStatement, dn, serialHex)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete certificate with DN "+dn+" and serial "+serial+" from database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete certificate with DN "+dn+" and serial "+serial+" from database")
		return err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	return nil
}

func makeOpenSSLTime(t time.Time) string {
	y := (int(t.Year()) % 100)
	validDate := fmt.Sprintf("%02d%02d%02d%02d%02d%02dZ", y, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return validDate
}
