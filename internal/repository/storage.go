package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"time"
)

var ErrNotFound = errors.New("timestamp not found")

type TimestampStorage interface {
	Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error)

	List(
		ctx context.Context,
		limit, offset int,
		externalID, tag, stage string,
		timestampFrom, timestampTo *time.Time,
		metaFilter map[string]any,
	) ([]*entity.Timestamp, error)

	Delete(ctx context.Context, id uuid.UUID) error
}
