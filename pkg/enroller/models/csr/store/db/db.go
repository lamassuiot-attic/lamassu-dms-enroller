package db

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/lamassuiot/enroller/pkg/enroller/models/csr"
	"github.com/lamassuiot/enroller/pkg/enroller/models/csr/store"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	_ "github.com/lib/pq"
)

func NewDB(driverName string, dataSourceName string, logger log.Logger) (store.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
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

func (db *DB) Insert(c csr.CSR) (int, error) {
	id := 0
	sqlStatement := `
	INSERT INTO csr_store(name, country, state, locality, organization, organization_unit, email, common_name, status, csrPath, url)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id;
	`
	err := db.QueryRow(sqlStatement, c.Name, c.CountryName, c.StateOrProvinceName, c.LocalityName, c.OrganizationName, c.OrganizationalUnitName, c.EmailAddress, c.CommonName, c.Status, c.CsrFilePath, c.Url).Scan(&id)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not insert CSR with CN "+c.CommonName+" in database")
		return -1, err
	}
	level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" inserted in database")
	return id, nil
}

func (db *DB) SelectAll() csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store;
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain CSRs from database or the database is empty")
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.Name, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
		if err != nil {
			level.Error(db.logger).Log("err", err, "msg", "Unable to read database CSR row")
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(c.Id)+" read from database")
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		level.Error(db.logger).Log("err", err)
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	level.Info(db.logger).Log("msg", strconv.Itoa(len(csrs))+" CSRs read from database")
	return csr.CSRs{CSRs: csrs}
}

func (db *DB) SelectAllByCN(cn string) csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store
	WHERE common_name = $1;
	`
	rows, err := db.Query(sqlStatement, cn)
	if err != nil {
		level.Error(db.logger).Log("err", err.Error, "msg", "Could not obtain CSR from database for CN "+cn+" or the database is empty")
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.Name, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath, &c.Url)
		if err != nil {
			level.Error(db.logger).Log("err", err, "msg", "Unable to read database CSR for CN "+cn)
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(c.Id)+" read from database")
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		level.Error(db.logger).Log("err", err)
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	level.Info(db.logger).Log("msg", strconv.Itoa(len(csrs))+" CSRs read from database for CN "+cn)
	return csr.CSRs{CSRs: csrs}
}

func (db *DB) SelectByStatus(status string) csr.CSRs {
	sqlStatement := `
	SELECT * 
	FROM csr_store;
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain CSR from database with status "+status)
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	defer rows.Close()
	csrs := make([]csr.CSR, 0)

	for rows.Next() {
		var c csr.CSR
		err := rows.Scan(&c.Id, &c.Name, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
		if err != nil {
			level.Error(db.logger).Log("err", err, "msg", "Unable to read database CSR with status "+status)
			return csr.CSRs{CSRs: []csr.CSR{}}
		}
		level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(c.Id)+" read from database")
		csrs = append(csrs, c)
	}
	if err = rows.Err(); err != nil {
		level.Error(db.logger).Log("err", err)
		return csr.CSRs{CSRs: []csr.CSR{}}
	}
	level.Info(db.logger).Log("msg", strconv.Itoa(len(csrs))+" CSRs read from database with status "+status)
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
	err := row.Scan(&c.Id, &c.Name, &c.CountryName, &c.StateOrProvinceName, &c.LocalityName, &c.OrganizationName, &c.OrganizationalUnitName, &c.CommonName, &c.EmailAddress, &c.Status, &c.CsrFilePath)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain CSR with ID "+strconv.Itoa(id)+" from database")
		return csr.CSR{}, err
	}
	level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" obtained from database")
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
		level.Error(db.logger).Log("err", err, "msg", "Could not update CSR with ID "+strconv.Itoa(id)+" status to "+c.Status)
		return csr.CSR{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not update CSR with ID "+strconv.Itoa(id)+" status to "+c.Status)
		return csr.CSR{}, err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return csr.CSR{}, err
	}
	level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(id)+" status updated to"+c.Status)
	return csr.CSR{}, nil
}

func (db *DB) UpdateFilePath(c csr.CSR) error {
	sqlStatement := `
	UPDATE csr_store
	SET csrPath = $1
	WHERE id = $2;
	`
	res, err := db.Exec(sqlStatement, c.CsrFilePath, c.Id)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not updated CSR with ID "+strconv.Itoa(c.Id)+" file path to "+c.CsrFilePath)
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not updated CSR with ID "+strconv.Itoa(c.Id)+" file path to "+c.CsrFilePath)
		return err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	level.Info(db.logger).Log("msg", "CSR with ID "+strconv.Itoa(c.Id)+" file path updated to "+c.CsrFilePath)
	return nil
}

func (db *DB) Delete(id int) error {
	sqlStatement := `
	DELETE FROM csr_store
	WHERE id = $1;
	`
	res, err := db.Exec(sqlStatement, id)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete CSR with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete CSR with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	return nil
}
