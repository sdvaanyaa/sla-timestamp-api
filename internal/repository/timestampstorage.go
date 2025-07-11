package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
)

var ErrNotFound = errors.New("timestamp not found")

type TimestampStorage interface {
	Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Timestamp, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
