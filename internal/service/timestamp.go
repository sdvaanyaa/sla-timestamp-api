package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache/rdscache"
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
	cache   *rdscache.Client
}

func New(storage repository.TimestampStorage, val *validator.Validate, cache *rdscache.Client) TimestampService {
	return &timestampService{
		storage: storage,
		val:     val,
		cache:   cache,
	}
}

func (s *timestampService) Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error) {
	if err := s.val.Struct(ts); err != nil {
		return uuid.Nil, ErrInvalidInput
	}

	_ = s.cache.Delete(ctx, "timestamps:list:*")

	return s.storage.Create(ctx, ts)
}

func (s *timestampService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	key := fmt.Sprintf("timestamp:%s", id.String())
	var ts entity.Timestamp
	if err := s.cache.Get(ctx, key, &ts); err == nil {
		return &ts, nil
	}

	tsPtr, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, tsPtr, 5*time.Minute)

	return tsPtr, nil
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

	paramsJSON, _ := json.Marshal(params)
	key := fmt.Sprintf("timestamps:list:%x", sha256.Sum256(paramsJSON))
	var list []*entity.Timestamp
	if err := s.cache.Get(ctx, key, &list); err == nil {
		return list, nil
	}

	list, err := s.storage.List(ctx, limit, offset, externalID, tag, stage, timestampFrom, timestampTo, metaFilter)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, list, 5*time.Minute)

	return list, nil
}

func (s *timestampService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("timestamp:%s", id.String())
	_ = s.cache.Delete(ctx, key)
	_ = s.cache.Delete(ctx, "timestamps:list:*")

	return nil
}
