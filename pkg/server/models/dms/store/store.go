package store

import (
	"context"

	"github.com/lamassuiot/lamassu-dms-enroller/pkg/server/models/dms"
)

type DB interface {
	Insert(ctx context.Context, d dms.DMS) (string, error)
	SelectAll(ctx context.Context) ([]dms.DMS, error)
	SelectByID(ctx context.Context, id string) (dms.DMS, error)
	SelectBySerialNumber(ctx context.Context, SerialNumber string) (string, error)
	UpdateByID(ctx context.Context, id string, status string, serialNumber string, encodedCsr string) (dms.DMS, error)
	Delete(ctx context.Context, id string) error
	InsertAuthorizedCAs(ctx context.Context, dmsid string, CAs []string) error
	DeleteAuthorizedCAs(ctx context.Context, dmsid string) error
	SelectByDMSIDAuthorizedCAs(ctx context.Context, dmsid string) ([]dms.AuthorizedCAs, error)
	SelectAllAuthorizedCAs(ctx context.Context) ([]dms.AuthorizedCAs, error)
}
