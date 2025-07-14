package service

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"log/slog"
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

	event := map[string]any{"action": "create", "data": ts}
	msg, err := json.Marshal(event)
	if err != nil {
		slog.Error("marshal event failed", slog.Any("error", err))
		return id, nil
	}
	_ = s.broker.Publish(ctx, msg)

	return id, nil
}
