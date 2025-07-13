package postgres

import (
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/pgdb"
)

type pgStorage struct {
	db *pgdb.Client
}

func New(db *pgdb.Client) repository.TimestampStorage {
	return &pgStorage{
		db: db,
	}
}
