package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	dmserrors "github.com/lamassuiot/dms-enroller/pkg/server/api/errors"
	"github.com/lamassuiot/dms-enroller/pkg/server/models/dms"
	"github.com/lamassuiot/dms-enroller/pkg/server/models/dms/store"
	"github.com/lamassuiot/dms-enroller/pkg/server/utils"
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

func (db *DB) Insert(ctx context.Context, d dms.DMS) (int, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	id := 0
	sqlStatement := `
	INSERT INTO dms_store(name, serialNumber, keyType, keyBits, csrBase64, status)
	VALUES($1, $2, $3, $4, $5, $6)
	RETURNING id;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: insert DMS with name "+d.Name+" in database", opentracing.ChildOf(parentSpan.Context()))
	err := db.QueryRow(sqlStatement, d.Name, d.SerialNumber, d.KeyMetadata.KeyType, d.KeyMetadata.KeyBits, d.CsrBase64, d.Status).Scan(&id)
	span.Finish()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not insert DMS with name "+d.Name+" in database")
		duplicationErr := &dmserrors.DuplicateResourceError{
			ResourceType: "DMS",
			ResourceId:   strconv.Itoa(id),
		}
		return -1, duplicationErr
	}
	level.Info(db.logger).Log("msg", "DMS with ID "+strconv.Itoa(id)+" inserted in database")
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
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain DMSs from database or the database is empty")
		return []dms.DMS{}, err
	}
	defer rows.Close()
	dmss := make([]dms.DMS, 0)

	for rows.Next() {
		var d dms.DMS
		err := rows.Scan(&d.Id, &d.Name, &d.SerialNumber, &d.KeyMetadata.KeyType, &d.KeyMetadata.KeyBits, &d.CsrBase64, &d.Status)
		if err != nil {
			level.Error(db.logger).Log("err", err, "msg", "Unable to read database DMS row")
			return []dms.DMS{}, err
		}
		level.Info(db.logger).Log("msg", "DMS with ID "+strconv.Itoa(d.Id)+" read from database")
		dmss = append(dmss, d)
	}
	if err = rows.Err(); err != nil {
		level.Error(db.logger).Log("err", err)
		return []dms.DMS{}, err
	}
	level.Info(db.logger).Log("msg", strconv.Itoa(len(dmss))+" DMSs read from database")
	return dmss, nil
}

func (db *DB) SelectByID(ctx context.Context, id int) (dms.DMS, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	SELECT *
	FROM dms_store
	WHERE id = $1;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: obtain DMS with ID "+strconv.Itoa(id)+" from database", opentracing.ChildOf(parentSpan.Context()))
	row := db.QueryRow(sqlStatement, id)
	span.Finish()
	var d dms.DMS
	err := row.Scan(&d.Id, &d.Name, &d.SerialNumber, &d.KeyMetadata.KeyType, &d.KeyMetadata.KeyBits, &d.CsrBase64, &d.Status)
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not obtain DMS with ID "+strconv.Itoa(id)+" from database")
		notFoundErr := &dmserrors.ResourceNotFoundError{
			ResourceType: "DMS",
			ResourceId:   strconv.Itoa(id),
		}
		return dms.DMS{}, notFoundErr
	}
	level.Info(db.logger).Log("msg", "DMS with ID "+strconv.Itoa(id)+" obtained from database")
	return d, nil
}

func (db *DB) UpdateByID(ctx context.Context, id int, status string, serialNumber string, encodedCsr string) (dms.DMS, error) {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	UPDATE dms_store
	SET status = $1, serialNumber = $2, csrBase64 = $3
	WHERE id = $4;
	`
	span := opentracing.StartSpan("lamassu-dms-enroller: update DMS with ID "+strconv.Itoa(id)+" status to "+status, opentracing.ChildOf(parentSpan.Context()))
	res, err := db.Exec(sqlStatement, status, serialNumber, encodedCsr, id)
	span.Finish()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not update DMS with ID "+strconv.Itoa(id)+" status to "+status)
		return dms.DMS{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not update DMS with ID "+strconv.Itoa(id)+" status to "+status)
		return dms.DMS{}, err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return dms.DMS{}, err
	}
	level.Info(db.logger).Log("msg", "DMS with ID "+strconv.Itoa(id)+" status updated to"+status)
	return db.SelectByID(ctx, id)
}

func (db *DB) Delete(ctx context.Context, id int) error {
	db.logger = ctx.Value(utils.LamassuLoggerContextKey).(log.Logger)
	parentSpan := opentracing.SpanFromContext(ctx)
	sqlStatement := `
	DELETE FROM dms_store
	WHERE id = $1;
	`
	span := opentracing.StartSpan("delete DMS with ID "+strconv.Itoa(id)+" from database", opentracing.ChildOf(parentSpan.Context()))
	res, err := db.Exec(sqlStatement, id)
	span.Finish()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete DMS with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		level.Error(db.logger).Log("err", err, "msg", "Could not delete DMS with ID "+strconv.Itoa(id)+" from database")
		return err
	}
	if count <= 0 {
		err = errors.New("No rows have been updated in database")
		level.Error(db.logger).Log("err", err)
		return err
	}
	return nil
}
