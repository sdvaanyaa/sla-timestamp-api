package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
)

func (s *pgStorage) Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error) {
	query := `
		INSERT INTO timestamps (external_id, timestamp, tag, stage, meta)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id uuid.UUID
	err := s.db.QueryRow(ctx, query, ts.ExternalID, ts.Timestamp, ts.Tag, ts.Stage, ts.Meta).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create: %w", ErrQueryFailed)
	}

	return id, nil
}
