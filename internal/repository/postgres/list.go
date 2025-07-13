package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"strings"
	"time"
)

func (s *pgStorage) List(
	ctx context.Context,
	limit, offset int,
	externalID, tag, stage string,
	timestampFrom, timestampTo *time.Time,
	metaFilter map[string]any,
) ([]*entity.Timestamp, error) {
	query, args, err := buildListQuery(externalID, tag, stage, timestampFrom, timestampTo, limit, offset, metaFilter)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list: %w", ErrQueryFailed)
	}
	defer rows.Close()

	var list []*entity.Timestamp

	for rows.Next() {
		ts, scanErr := scanTimestampRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}

		list = append(list, ts)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("list: %w", ErrRowsFailed)
	}

	return list, nil
}

func buildListQuery(
	externalID, tag, stage string,
	timestampFrom, timestampTo *time.Time,
	limit, offset int,
	metaFilter map[string]any,
) (string, []any, error) {
	var query strings.Builder
	query.WriteString("SELECT id, external_id, timestamp, tag, stage, meta FROM timestamps WHERE 1=1")

	var args []any
	argIndex := 1

	if externalID != "" {
		query.WriteString(fmt.Sprintf(" AND external_id = $%d", argIndex))
		args = append(args, externalID)
		argIndex++
	}

	if tag != "" {
		query.WriteString(fmt.Sprintf(" AND tag = $%d", argIndex))
		args = append(args, tag)
		argIndex++
	}

	if stage != "" {
		query.WriteString(fmt.Sprintf(" AND stage = $%d", argIndex))
		args = append(args, stage)
		argIndex++
	}

	if timestampFrom != nil {
		query.WriteString(fmt.Sprintf(" AND timestamp >= $%d", argIndex))
		args = append(args, *timestampFrom)
		argIndex++
	}

	if timestampTo != nil {
		query.WriteString(fmt.Sprintf(" AND timestamp <= $%d", argIndex))
		args = append(args, *timestampTo)
		argIndex++
	}

	if len(metaFilter) > 0 {
		metaJSON, err := json.Marshal(metaFilter)
		if err != nil {
			return "", nil, fmt.Errorf("marshal meta_filter: %w", err)
		}
		query.WriteString(fmt.Sprintf(" AND meta @> $%d", argIndex))
		args = append(args, string(metaJSON))
		argIndex++
	}

	query.WriteString(fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
	args = append(args, limit, offset)

	return query.String(), args, nil
}

func scanTimestampRow(rows pgx.Rows) (*entity.Timestamp, error) {
	var ts entity.Timestamp
	var metaBytes []byte
	if err := rows.Scan(&ts.ID, &ts.ExternalID, &ts.Timestamp, &ts.Tag, &ts.Stage, &metaBytes); err != nil {
		return nil, fmt.Errorf("scan row: %w", ErrScanFailed)
	}
	if metaBytes != nil {
		if err := json.Unmarshal(metaBytes, &ts.Meta); err != nil {
			return nil, fmt.Errorf("unmarshal meta: %w", ErrUnmarshalFailed)
		}
	}
	return &ts, nil
}
