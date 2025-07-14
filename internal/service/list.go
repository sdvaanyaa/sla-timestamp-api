package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"time"
)

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

	if err := s.validateListParams(params); err != nil {
		return nil, err
	}

	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return s.storage.List(ctx, limit, offset, externalID, tag, stage, timestampFrom, timestampTo, metaFilter)
	}

	key := fmt.Sprintf("timestamps:list:%x", sha256.Sum256(paramsJSON))
	var list []*entity.Timestamp
	if err = s.cache.Get(ctx, key, &list); err == nil {
		return list, nil
	}

	list, err = s.storage.List(ctx, limit, offset, externalID, tag, stage, timestampFrom, timestampTo, metaFilter)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, key, list, CacheTTL)

	return list, nil
}

func (s *timestampService) validateListParams(params *entity.ListQueryParams) error {
	if err := s.val.Struct(params); err != nil {
		return ErrInvalidInput
	}

	if len(params.MetaFilter) > 0 {
		for k := range params.MetaFilter {
			if k == "" {
				return ErrInvalidInput
			}
		}
	}

	if params.TimestampFrom != nil && params.TimestampTo != nil && params.TimestampFrom.After(*params.TimestampTo) {
		return ErrInvalidInput
	}

	return nil
}
