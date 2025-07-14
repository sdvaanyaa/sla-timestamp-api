package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
)

func (s *timestampService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	key := fmt.Sprintf(TimestampCachePrefix, id.String())
	var ts entity.Timestamp
	if err := s.cache.Get(ctx, key, &ts); err == nil {
		return &ts, nil
	}

	tsPtr, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, tsPtr, CacheTTL)

	return tsPtr, nil
}
