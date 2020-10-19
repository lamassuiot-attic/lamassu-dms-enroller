package db

import (
	"database/sql"
	"enroller/pkg/enroller/crypto"
	"errors"
	"fmt"

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

func (db *DB) InsertCSR(csr crypto.CSR) (int, error) {
	id := 0
	sqlStatement := `
	INSERT INTO csr_store(countryName, stateOrProvinceName, localityName, organizationName, organizationalUnitName, emailAddress, commonName, status, csrFilePath)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id;
	`
	err := db.QueryRow(sqlStatement, csr.CountryName, csr.StateOrProvinceName, csr.LocalityName, csr.OrganizationName, csr.OrganizationalUnitName, csr.EmailAddress, csr.CommonName, csr.Status, csr.CsrFilePath).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (db *DB) SelectCSRsByStatus(status string) crypto.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store;
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return crypto.CSRs{CSRs: []crypto.CSR{}}
	}
	defer rows.Close()
	csrs := make([]crypto.CSR, 0)

	for rows.Next() {
		var csr crypto.CSR
		err := rows.Scan(&csr.Id, &csr.CountryName, &csr.StateOrProvinceName, &csr.LocalityName, &csr.OrganizationName, &csr.OrganizationalUnitName, &csr.CommonName, &csr.EmailAddress, &csr.Status, &csr.CsrFilePath)
		if err != nil {
			return crypto.CSRs{CSRs: []crypto.CSR{}}
		}
		csrs = append(csrs, csr)
	}
	if err = rows.Err(); err != nil {
		return crypto.CSRs{CSRs: []crypto.CSR{}}
	}
	return crypto.CSRs{CSRs: csrs}
}

func (db *DB) SelectCSRByID(id int) (crypto.CSR, error) {
	sqlStatement := `
	SELECT *
	FROM csr_store
	WHERE id = $1;
	`
	row := db.QueryRow(sqlStatement, id)
	var csr crypto.CSR
	err := row.Scan(&csr.Id, &csr.CountryName, &csr.StateOrProvinceName, &csr.LocalityName, &csr.OrganizationName, &csr.OrganizationalUnitName, &csr.CommonName, &csr.EmailAddress, &csr.Status, &csr.CsrFilePath)
	if err != nil {
		return crypto.CSR{}, err
	}
	return csr, nil
}

func (db *DB) UpdateCSRByID(id int, csr crypto.CSR) (crypto.CSR, error) {
	sqlStatement := `
	UPDATE csr_store
	SET status = $1
	WHERE id = $2;
	`
	res, err := db.Exec(sqlStatement, csr.Status, id)
	if err != nil {
		return crypto.CSR{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return crypto.CSR{}, err
	}
	if count <= 0 {
		return crypto.CSR{}, errors.New("No rows updated")
	}
	return crypto.CSR{}, nil
}

func (db *DB) UpdateCSRFilePath(csr crypto.CSR) error {
	sqlStatement := `
	UPDATE csr_store
	SET csrfilepath = $1
	WHERE id = $2;
	`
	res, err := db.Exec(sqlStatement, csr.CsrFilePath, csr.Id)
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

func (db *DB) DeleteCSR(id int) error {
	sqlStatement := `
	DELETE FROM csr_store
	WHERE id = $1;
	`
	res, err := db.Exec(sqlStatement, id)
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

// Function for Tests
func (db *DB) TruncateTable() error {
	_, err := db.Exec("TRUNCATE TABLE csr_store;")
	if err != nil {
		return err
	}
	return nil
}

// Function for Tests
func (db *DB) CloseDB() error {
	error := db.Close()
	return error
}
