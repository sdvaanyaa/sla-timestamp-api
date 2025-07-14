package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
)

func (s *timestampService) Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error) {
	if err := s.val.Struct(ts); err != nil {
		return uuid.Nil, ErrInvalidInput
	}

	id, err := s.storage.Create(ctx, ts)
	if err != nil {
		return uuid.Nil, err
	}

	ts.ID = id
	key := fmt.Sprintf(TimestampCachePrefix, id.String())
	_ = s.cache.Set(ctx, key, ts, CacheTTL)

	_ = s.cache.Delete(ctx, ListCachePrefix)

	return id, nil
}
