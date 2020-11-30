package db

import (
	"database/sql"
	"enroller/pkg/enroller/models/csr"
	"enroller/pkg/enroller/models/csr/store"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
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

func (db *DB) Insert(c csr.CSR) (int, error) {
	id := 0
	sqlStatement := `
	INSERT INTO csr_store(countryName, stateOrProvinceName, localityName, organizationName, organizationalUnitName, emailAddress, commonName, status, csrFilePath)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id;
	`
	err := db.QueryRow(sqlStatement, c.CountryName, c.StateOrProvinceName, c.LocalityName, c.OrganizationName, c.OrganizationalUnitName, c.EmailAddress, c.CommonName, c.Status, c.CsrFilePath).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (db *DB) SelectAll() csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store;
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
		if err != nil {
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	return csr.CSRs{CSRs: csrs}
}

func (db *DB) SelectAllByCN(cn string) csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store
	WHERE commonName = $1;
	`

	rows, err := db.Query(sqlStatement, cn)
	if err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
		if err != nil {
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	return csr.CSRs{CSRs: csrs}
}

func (db *DB) SelectByStatus(status string) csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store;
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
		if err != nil {
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	return csr.CSRs{CSRs: csrs}
}

func (db *DB) SelectByID(id int) (csr.CSR, error) {
	sqlStatement := `
	SELECT *
	FROM csr_store
	WHERE id = $1;
	`
	row := db.QueryRow(sqlStatement, id)
	var c csr.CSR
	err := row.Scan(&c.Id, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
	if err != nil {
		return csr.CSR{}, err
	}
	return c, nil
}

func (db *DB) UpdateByID(id int, c csr.CSR) (csr.CSR, error) {
	sqlStatement := `
	UPDATE csr_store
	SET status = $1
	WHERE id = $2;
	`
	res, err := db.Exec(sqlStatement, c.Status, id)
	if err != nil {
		return csr.CSR{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return csr.CSR{}, err
	}
	if count <= 0 {
		return csr.CSR{}, errors.New("No rows updated")
	}
	return csr.CSR{}, nil
}

func (db *DB) UpdateFilePath(c csr.CSR) error {
	sqlStatement := `
	UPDATE csr_store
	SET csrfilepath = $1
	WHERE id = $2;
	`
	res, err := db.Exec(sqlStatement, c.CsrFilePath, c.Id)
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

func (db *DB) Delete(id int) error {
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
