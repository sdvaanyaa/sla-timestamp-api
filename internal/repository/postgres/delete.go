package postgres

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
)

func (s *pgStorage) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM timestamps WHERE id = $1`

	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete: %w", ErrQueryFailed)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete: %w", repository.ErrNotFound)
	}

	return nil
}
