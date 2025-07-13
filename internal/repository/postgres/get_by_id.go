package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
)

func (s *pgStorage) GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error) {
	query := `
		SELECT id, external_id, timestamp, tag, stage, meta
		FROM timestamps
		WHERE id = $1
	`
	var ts entity.Timestamp
	var metaBytes []byte

	err := s.db.QueryRow(ctx, query, id).Scan(&ts.ID, &ts.ExternalID, &ts.Timestamp, &ts.Tag, &ts.Stage, &metaBytes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get by id: %w", repository.ErrNotFound)
		}
		return nil, fmt.Errorf("get by id: %w", ErrQueryFailed)
	}

	if metaBytes != nil {
		if err = json.Unmarshal(metaBytes, &ts.Meta); err != nil {
			return nil, fmt.Errorf("get by id: %w", ErrUnmarshalFailed)
		}
	}

	return &ts, nil
}
