package service

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"time"
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
}

func New(storage repository.TimestampStorage, val *validator.Validate) TimestampService {
	return &timestampService{
		storage: storage,
		val:     val,
	}
}

func (s *timestampService) Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error) {
	if err := s.val.Struct(ts); err != nil {
		return uuid.Nil, ErrInvalidInput
	}

	return s.storage.Create(ctx, ts)
}

func (s *timestampService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	return s.storage.GetByID(ctx, id)
}

func (s *timestampService) List(
	ctx context.Context,
	limit, offset int,
	externalID, tag, stage string,
	timestampFrom, timestampTo *time.Time,
	metaFilter map[string]any,
) ([]*entity.Timestamp, error) {
	params := &entity.ListQueryParams{
		Limit:         limit,
		Offset:        offset,
		ExternalID:    externalID,
		Tag:           tag,
		Stage:         stage,
		TimestampFrom: timestampFrom,
		TimestampTo:   timestampTo,
		MetaFilter:    metaFilter,
	}

	if err := s.val.Struct(params); err != nil {
		return nil, ErrInvalidInput
	}

	if len(metaFilter) > 0 {
		for k := range metaFilter {
			if k == "" {
				return nil, ErrInvalidInput
			}
		}
	}

	if timestampFrom != nil && timestampTo != nil && timestampFrom.After(*timestampTo) {
		return nil, ErrInvalidInput
	}

	return s.storage.List(ctx, limit, offset, externalID, tag, stage, timestampFrom, timestampTo, metaFilter)
}

func (s *timestampService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	return s.storage.Delete(ctx, id)
}
