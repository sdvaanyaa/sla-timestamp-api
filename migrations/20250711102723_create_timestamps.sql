-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE tag_enum AS ENUM ('incident', 'sla', 'deployment', 'maintenance', 'alert');
CREATE TYPE stage_enum AS ENUM ('created', 'acknowledged', 'in_progress', 'resolved', 'closed');

CREATE TABLE timestamps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    tag tag_enum NOT NULL,
    stage stage_enum NOT NULL,
    meta JSONB,
    CONSTRAINT unique_timestamp UNIQUE (external_id, tag, stage)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS timestamps;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS tag_enum;
DROP TYPE IF EXISTS stage_enum;
-- +goose StatementEnd
