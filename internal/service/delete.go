package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
)

func (s *timestampService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}

	key := fmt.Sprintf(TimestampCachePrefix, id.String())
	_ = s.cache.Delete(ctx, key)
	_ = s.cache.Delete(ctx, ListCachePrefix)

	return nil
}
