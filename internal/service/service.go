package service

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache"
	"time"
)

const (
	CacheTTL             = 5 * time.Minute
	ListCachePrefix      = "timestamps:list:*"
	TimestampCachePrefix = "timestamp:%s"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

type TimestampService interface {
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

type timestampService struct {
	storage repository.TimestampStorage
	val     *validator.Validate
	cache   cache.Cache
}

func New(storage repository.TimestampStorage, val *validator.Validate, cache cache.Cache) TimestampService {
	return &timestampService{
		storage: storage,
		val:     val,
		cache:   cache,
	}
}
