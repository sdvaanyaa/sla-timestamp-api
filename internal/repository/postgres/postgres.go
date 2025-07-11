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
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/pgdb"
)

type pgStorage struct {
	db *pgdb.Client
}

func New(db *pgdb.Client) repository.TimestampStorage {
	return &pgStorage{
		db: db,
	}
}

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

func (s *pgStorage) List(ctx context.Context, limit, offset int) ([]*entity.Timestamp, error) {
	query := `
		SELECT id, external_id, timestamp, tag, stage, meta
		FROM timestamps
		ORDER BY timestamp DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list: %w", ErrQueryFailed)
	}
	defer rows.Close()

	var list []*entity.Timestamp

	for rows.Next() {
		var ts entity.Timestamp
		var metaBytes []byte

		if err = rows.Scan(&ts.ID, &ts.ExternalID, &ts.Timestamp, &ts.Tag, &ts.Stage, &metaBytes); err != nil {
			return nil, fmt.Errorf("list: %w", ErrScanFailed)
		}

		if metaBytes != nil {
			if err = json.Unmarshal(metaBytes, &ts.Meta); err != nil {
				return nil, fmt.Errorf("list: %w", ErrUnmarshalFailed)
			}
		}

		list = append(list, &ts)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("list: %w", ErrRowsFailed)
	}

	return list, nil
}

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
