package integration

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/config"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository/postgres"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
)

type TimestampRepoSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer testcontainers.Container
	client      *pgdb.Client
	repo        repository.TimestampStorage
}

func (s *TimestampRepoSuite) SetupSuite() {
	s.ctx, s.pgContainer, s.client = setupPostgresContainer(s.T())
	s.repo = postgres.New(s.client)

	schema := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TYPE tag_enum AS ENUM ('incident', 'sla', 'deployment', 'maintenance', 'alert');
		CREATE TYPE stage_enum AS ENUM ('created', 'acknowledged', 'in_progress', 'resolved', 'closed');
		CREATE TABLE IF NOT EXISTS timestamps (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			external_id VARCHAR(255) NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			tag tag_enum NOT NULL,
			stage stage_enum NOT NULL,
			meta JSONB,
			CONSTRAINT unique_timestamp UNIQUE (external_id, tag, stage)
		);
	`
	_, err := s.client.Exec(s.ctx, schema)
	require.NoError(s.T(), err)
}

func (s *TimestampRepoSuite) SetupTest() {
	_, err := s.client.Exec(s.ctx, "TRUNCATE TABLE timestamps RESTART IDENTITY CASCADE")
	require.NoError(s.T(), err)
}

func (s *TimestampRepoSuite) TearDownSuite() {
	s.client.Close()
	err := s.pgContainer.Terminate(s.ctx)
	require.NoError(s.T(), err)
}

func (s *TimestampRepoSuite) TestCreate() {
	tests := []struct {
		name    string
		ts      *entity.Timestamp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			ts: &entity.Timestamp{
				ExternalID: "test-external",
				Timestamp:  time.Now().UTC(),
				Tag:        entity.TagIncident,
				Stage:      entity.StageCreated,
				Meta:       map[string]any{"key": "value"},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Tag",
			ts: &entity.Timestamp{
				ExternalID: "test-external",
				Timestamp:  time.Now().UTC(),
				Tag:        "invalid",
				Stage:      entity.StageCreated,
			},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Parallel()

			id, err := s.repo.Create(s.ctx, tt.ts)
			tt.wantErr(s.T(), err)
			if err != nil {
				return
			}
			assert.NotEqual(s.T(), uuid.Nil, id)

			var stored entity.Timestamp
			var metaBytes []byte
			query := `SELECT id, external_id, timestamp, tag, stage, meta FROM timestamps WHERE id = $1`
			err = s.client.QueryRow(s.ctx, query, id).Scan(
				&stored.ID,
				&stored.ExternalID,
				&stored.Timestamp,
				&stored.Tag,
				&stored.Stage,
				&metaBytes,
			)

			assert.NoError(s.T(), err)

			if tt.ts.Meta != nil {
				err = json.Unmarshal(metaBytes, &stored.Meta)
				assert.NoError(s.T(), err)
			}

			assertApproxEqualTimestamp(s.T(), tt.ts, &stored)
		})
	}
}

func (s *TimestampRepoSuite) TestGetByID() {
	ts := &entity.Timestamp{
		ExternalID: "test-external",
		Timestamp:  time.Now().UTC(),
		Tag:        entity.TagIncident,
		Stage:      entity.StageCreated,
		Meta:       map[string]any{"key": "value"},
	}
	id, err := s.repo.Create(s.ctx, ts)
	require.NoError(s.T(), err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *entity.Timestamp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Success",
			id:      id,
			want:    ts,
			wantErr: assert.NoError,
		},
		{
			name:    "Not Found",
			id:      uuid.New(),
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.T().Parallel()

			got, errGet := s.repo.GetByID(s.ctx, tt.id)
			tt.wantErr(s.T(), errGet)
			if errGet == nil {
				assertApproxEqualTimestamp(s.T(), tt.want, got)
			}
		})
	}
}

func TestStorageSuite(t *testing.T) {
	suite.Run(t, new(TimestampRepoSuite))
}

func setupPostgresContainer(t *testing.T) (context.Context, testcontainers.Container, *pgdb.Client) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "sla_test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	cfg := config.PostgresConfig{
		Host:     host,
		Port:     port.Port(),
		Username: "postgres",
		Password: "postgres",
		Database: "sla_test",
		SSLMode:  "disable",
	}
	client, err := pgdb.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))

	require.NoError(t, err)

	return ctx, pgContainer, client
}

func assertApproxEqualTimestamp(t *testing.T, want, got *entity.Timestamp, msgAndArgs ...any) {
	t.Helper()

	opts := []cmp.Option{
		cmpopts.EquateApproxTime(time.Second),
		cmpopts.IgnoreFields(entity.Timestamp{}, "ID"),
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		assert.Fail(t, "mismatch (-want +got):\n"+diff, msgAndArgs...)
	}
}
