package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	dmserrors "github.com/lamassuiot/lamassu-dms-enroller/pkg/server/api/errors"
	"github.com/lamassuiot/lamassu-dms-enroller/pkg/server/models/dms"
	"github.com/lamassuiot/lamassu-dms-enroller/pkg/server/models/dms/store"
	"github.com/lamassuiot/lamassu-dms-enroller/pkg/server/utils"
	"github.com/opentracing/opentracing-go"

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
		time.Sleep(5 * time.Second)
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

func (db *DB) Insert(ctx context.Context, d dms.DMS) (string, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	var id string
	sqlStatement := `
	INSERT INTO dms_store(id,name, serialNumber, keyType, keyBits, csrBase64, status, creation_ts, modification_ts)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: insert DMS with name "+d.Name+" in database", opentracing.ChildOf(parentSpan.Context()))
	err := db.QueryRow(sqlStatement, d.Id, d.Name, d.SerialNumber, d.KeyMetadata.KeyType, d.KeyMetadata.KeyBits, d.CsrBase64, d.Status, time.Now(), time.Now()).Scan(&id)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not insert DMS with name "+d.Name+" in database")
		duplicationErr := &dmserrors.DuplicateResourceError{
			ResourceType: "DMS",
			ResourceId:   id,
		}
		return "", duplicationErr
	}
	level.Debug(db.logger).Log("msg", "DMS with ID "+id+" inserted in database")
	return id, nil
}

func (db *DB) SelectAll(ctx context.Context) ([]dms.DMS, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT * 
	FROM dms_store;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain DMSs from database", opentracing.ChildOf(parentSpan.Context()))
	rows, err := db.Query(sqlStatement)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not obtain DMSs from database or the database is empty")
		return []dms.DMS{}, err
	}
	defer rows.Close()
	dmss := make([]dms.DMS, 0)

	for rows.Next() {
		var d dms.DMS
		err := rows.Scan(&d.Id, &d.Name, &d.SerialNumber, &d.KeyMetadata.KeyType, &d.KeyMetadata.KeyBits, &d.CsrBase64, &d.Status, &d.CreationTimestamp, &d.ModificationTimestamp)
		if err != nil {
			return []dms.DMS{}, err
		}
		d.KeyMetadata.KeyStrength = getKeyStrength(d.KeyMetadata.KeyType, d.KeyMetadata.KeyBits)
		dmss = append(dmss, d)
	}
	if err = rows.Err(); err != nil {
		level.Debug(db.logger).Log("err", err)
		return []dms.DMS{}, err
	}
	level.Debug(db.logger).Log("msg", strconv.Itoa(len(dmss))+" DMSs read from database")
	return dmss, nil
}

func (db *DB) SelectByID(ctx context.Context, id string) (dms.DMS, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT *
	FROM dms_store
	WHERE id = $1;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain DMS with ID "+id+" from database", opentracing.ChildOf(parentSpan.Context()))
	row := db.QueryRow(sqlStatement, id)
	span.Finish()
	var d dms.DMS
	err := row.Scan(&d.Id, &d.Name, &d.SerialNumber, &d.KeyMetadata.KeyType, &d.KeyMetadata.KeyBits, &d.CsrBase64, &d.Status, &d.CreationTimestamp, &d.ModificationTimestamp)
	if err != nil {
		notFoundErr := &dmserrors.ResourceNotFoundError{
			ResourceType: "DMS",
			ResourceId:   id,
		}
		return dms.DMS{}, notFoundErr
	}
	d.KeyMetadata.KeyStrength = getKeyStrength(d.KeyMetadata.KeyType, d.KeyMetadata.KeyBits)
	level.Debug(db.logger).Log("msg", "DMS with ID "+id+" obtained from database")
	return d, nil
}
func (db *DB) SelectBySerialNumber(ctx context.Context, SerialNumber string) (string, error) {
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT *
	FROM dms_store
	WHERE serialNumber = $1;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain DMS with SerialNumber "+SerialNumber+" from database", opentracing.ChildOf(parentSpan.Context()))
	row, err := db.Query(sqlStatement, SerialNumber)
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not obtain DMS")
		return "", err
	}
	span.Finish()
	defer row.Close()
	var d dms.DMS
	for row.Next() {

		err = row.Scan(&d.Id, &d.Name, &d.SerialNumber, &d.KeyMetadata.KeyType, &d.KeyMetadata.KeyBits, &d.CsrBase64, &d.Status, &d.CreationTimestamp, &d.ModificationTimestamp)
		if err != nil {
			return "", err
		}
	}
	return d.Id, nil
}

func (db *DB) UpdateByID(ctx context.Context, id string, status string, serialNumber string, encodedCsr string) (dms.DMS, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	UPDATE dms_store
	SET status = $1, serialNumber = $2, csrBase64 = $3, modification_ts = $4
	WHERE id = $5;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: update DMS with ID "+id+" status to "+status, opentracing.ChildOf(parentSpan.Context()))
	res, err := db.Exec(sqlStatement, status, serialNumber, encodedCsr, time.Now(), id)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not update DMS with ID "+id+" status to "+status)
		return dms.DMS{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return dms.DMS{}, err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Debug(db.logger).Log("err", err)
		return dms.DMS{}, err
	}
	level.Debug(db.logger).Log("msg", "DMS with ID "+id+" status updated to"+status)
	return db.SelectByID(ctx, id)
}

func (db *DB) Delete(ctx context.Context, id string) error {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	DELETE FROM dms_store
	WHERE id = $1;
	`
	span := opentracing.StartSpan("delete DMS with ID "+id+" from database", opentracing.ChildOf(parentSpan.Context()))
	res, err := db.Exec(sqlStatement, id)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not delete DMS with ID "+id+" from database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		err = errors.New("no rows have been updated in database")
		level.Debug(db.logger).Log("err", err)
		return err
	}
	return nil
}

func (db *DB) InsertAuthorizedCAs(ctx context.Context, dmsid string, CAs []string) error {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	for i := 0; i < len(CAs); i++ {
		sqlStatement := `
		INSERT INTO authorized_cas(dmsid,caname)
		VALUES($1,$2)
		RETURNING dmsid;
		`
		span := opentracing.StartSpan("lamassu-dms-enroller: insert Authorized CA with name "+CAs[i]+" in authorized_cas database", opentracing.ChildOf(parentSpan.Context()))
		err := db.QueryRow(sqlStatement, dmsid, CAs[i]).Scan(&dmsid)
		span.Finish()
		if err != nil {
			level.Debug(db.logger).Log("err", err, "msg", "Could not insert CA with name "+CAs[i]+" in authorized_cas database")
			duplicationErr := &dmserrors.DuplicateResourceError{
				ResourceType: "DMS",
				ResourceId:   dmsid,
			}
			return duplicationErr
		}
	}
	level.Debug(db.logger).Log("msg", "DMS with ID "+dmsid+" inserted in authorized_cas database")
	return nil
}
func (db *DB) DeleteAuthorizedCAs(ctx context.Context, dmsid string) error {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	DELETE FROM authorized_cas
	WHERE dmsid = $1;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: delete DMS with ID "+dmsid+" from authorized_cas database", opentracing.ChildOf(parentSpan.Context()))
	res, err := db.Exec(sqlStatement, dmsid)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not delete DMS with ID "+dmsid+" in authorized_cas database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count <= 0 {
		err = errors.New("no rows have been updated in database")
		level.Debug(db.logger).Log("err", err)
		return err
	}
	level.Debug(db.logger).Log("msg", "DMS with ID "+dmsid+" deleted in authorized_cas database")
	return nil
}

func (db *DB) SelectByDMSIDAuthorizedCAs(ctx context.Context, dmsid string) ([]dms.AuthorizedCAs, error) {
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT * 
	FROM authorized_cas
	WHERE dmsid = $1;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain Authorized CAs with DMS ID"+dmsid+"from database", opentracing.ChildOf(parentSpan.Context()))
	rows, err := db.Query(sqlStatement, dmsid)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not obtain DMSs from database or the database is empty")
		return nil, err
	}
	defer rows.Close()
	cass := make([]dms.AuthorizedCAs, 0)

	for rows.Next() {
		var d dms.AuthorizedCAs
		err := rows.Scan(&d.DmsId, &d.CaName)
		if err != nil {
			return nil, err
		}
		cass = append(cass, d)
	}
	if err = rows.Err(); err != nil {
		level.Debug(db.logger).Log("err", err)
		return nil, err
	}
	return cass, nil
}
func (db *DB) SelectAllAuthorizedCAs(ctx context.Context) ([]dms.AuthorizedCAs, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT * 
	FROM authorized_cas;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain authorized CAs from database", opentracing.ChildOf(parentSpan.Context()))
	rows, err := db.Query(sqlStatement)
	span.Finish()
	if err != nil {
		level.Debug(db.logger).Log("err", err, "msg", "Could not obtain authorized CAs from database or the database is empty")
		return []dms.AuthorizedCAs{}, err
	}
	defer rows.Close()
	dmss := make([]dms.AuthorizedCAs, 0)

	for rows.Next() {
		var d dms.AuthorizedCAs
		err := rows.Scan(&d.DmsId, &d.CaName)
		if err != nil {
			return []dms.AuthorizedCAs{}, err
		}
		dmss = append(dmss, d)
	}
	if err = rows.Err(); err != nil {
		level.Debug(db.logger).Log("err", err)
		return []dms.AuthorizedCAs{}, err
	}
	level.Debug(db.logger).Log("msg", strconv.Itoa(len(dmss))+" DMSs read from database")
	return dmss, nil
}

func getKeyStrength(keyType string, keyBits int) string {
	var keyStrength string = "unknown"
	switch keyType {
	case "RSA":
		if keyBits < 2048 {
			keyStrength = "low"
		} else if keyBits >= 2048 && keyBits < 3072 {
			keyStrength = "medium"
		} else {
			keyStrength = "high"
		}
	case "EC":
		if keyBits < 224 {
			keyStrength = "low"
		} else if keyBits >= 224 && keyBits < 256 {
			keyStrength = "medium"
		} else {
			keyStrength = "high"
		}
	}
	return keyStrength
}
