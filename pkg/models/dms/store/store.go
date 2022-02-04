package store

import (
	"context"

	"github.com/lamassuiot/dms-enroller/pkg/models/dms"
)

type DB interface {
	Insert(ctx context.Context, d dms.DMS) (int, error)
	SelectAll(ctx context.Context) ([]dms.DMS, error)
	SelectByID(ctx context.Context, id int) (dms.DMS, error)
	UpdateByID(ctx context.Context, id int, status string, serialNumber string, encodedCsr string) (dms.DMS, error)
	Delete(ctx context.Context, id int) error
}
