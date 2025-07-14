package service

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
)

func (s *timestampService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	err := s.storage.Delete(ctx, id)
	if err != nil {
		return err
	}

	event := map[string]any{"action": "delete", "id": id.String()}
	msg, err := json.Marshal(event)
	if err != nil {
		slog.Error("marshal event failed", slog.Any("error", err))
		return nil
	}

	_ = s.broker.Publish(ctx, msg)

	return nil
}
