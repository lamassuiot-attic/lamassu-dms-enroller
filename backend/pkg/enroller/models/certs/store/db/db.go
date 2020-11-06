package db

import (
	"database/sql"
	"enroller/pkg/enroller/models/certs"
	"enroller/pkg/enroller/models/certs/store"
	"errors"
	"fmt"
	"math/big"
)

func NewDB(driverName string, dataSourceName string) (store.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	err = checkDBAlive(db)
	for err != nil {
		fmt.Println("Trying to connect to DB")
		err = checkDBAlive(db)
	}

	return &DB{db}, nil
}

type DB struct {
	*sql.DB
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
		return err
	}
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
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected <= 0 {
		return errors.New("No rows updated")
	}
	return nil

}
