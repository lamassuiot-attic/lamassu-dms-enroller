package db

import (
	"database/sql"
	"enroller/pkg/scep/crypto"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func NewDB(driverName string, dataSourceName string) (*DB, error) {
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
		return err
	}
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
		return crypto.CRT{}, err
	}
	return crt, nil
}

func (db *DB) GetCRTs() (crypto.CRTs, error) {
	sqlStatement := `
	SELECT *
	FROM ca_store;
	`

	rows, err := db.Query(sqlStatement)

	if err != nil {
		return crypto.CRTs{CRTs: []crypto.CRT{}}, err
	}

	defer rows.Close()
	crts := make([]crypto.CRT, 0)

	for rows.Next() {
		var crt crypto.CRT
		err := rows.Scan(&crt.Status, &crt.ExpirationDate, &crt.RevocationDate, &crt.Serial, &crt.DN, &crt.CRTPath, &crt.Key, &crt.KeySize)
		if err != nil {
			return crypto.CRTs{CRTs: []crypto.CRT{}}, err
		}
		crts = append(crts, crt)
	}

	if err = rows.Err(); err != nil {
		return crypto.CRTs{CRTs: []crypto.CRT{}}, err
	}

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
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count <= 0 {
		return errors.New("No rows updated")
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
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		return errors.New("No updates")
	}
	return nil
}

func makeOpenSSLTime(t time.Time) string {
	y := (int(t.Year()) % 100)
	validDate := fmt.Sprintf("%02d%02d%02d%02d%02d%02dZ", y, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return validDate
}
