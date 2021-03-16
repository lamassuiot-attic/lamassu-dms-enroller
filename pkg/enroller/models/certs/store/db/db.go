package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/lamassuiot/enroller/pkg/enroller/models/certs"
	"github.com/lamassuiot/enroller/pkg/enroller/models/certs/store"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"math/big"
)

func NewDB(driverName string, dataSourceName string, logger log.Logger) (store.DB, error) {
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

func (db *DB) Insert(crt certs.CRT) error {
	sqlStatement := `

	INSERT INTO ca_store(id, status, expirationDate, revocationDate, serial, dn, certPath)
	VALUES($1, $2, $3, $4, $5, $6, $7)
	RETURNING serial;
	`
	serialHex := fmt.Sprintf("%x", crt.Serial)
	var serial string

	err := db.QueryRow(sqlStatement, crt.ID, crt.Status, crt.ExpirationDate, crt.RevocationDate, serialHex, crt.DN, crt.CertPath).Scan(&serial)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not insert certificate with ID "+strconv.Itoa(crt.ID)+" in database")
		return err
	}
	level.Info(db.logger).Log("msg", "Certificate with ID "+strconv.Itoa(crt.ID)+" inserted in database")
	return nil
}

func (db *DB) Serial() (*big.Int, error) {
	var serial string

	sqlStatement := `
	SELECT serial
	FROM ca_store
	ORDER BY serial DESC
	LIMIT 1;
	`
	row := db.QueryRow(sqlStatement)
	err := row.Scan(&serial)

	if err != nil {
		return big.NewInt(2), nil
	}

	s, _ := new(big.Int).SetString(serial, 16)
	s = s.Add(s, big.NewInt(1))
	return s, nil
}

func (db *DB) Revoke(id int, revocationDate string) error {
	sqlStatement := `
	UPDATE ca_store
	SET status = 'R', revocationDate = $1
	WHERE id = $2;
	`

	res, err := db.Exec(sqlStatement, revocationDate, id)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not revoke certificate with ID "+strconv.Itoa(id)+" in database")
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not revoke certificate with ID "+strconv.Itoa(id)+" in database")
		return err
	}

	if rowsAffected <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	return nil
}

func (db *DB) Delete(id int) error {
	sqlStatement := `
	DELETE FROM ca_store
	WHERE id = $1;
	`
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete certificate with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete certificate with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	return nil
}
