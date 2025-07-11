-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE timestamps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    tag VARCHAR(50) NOT NULL,
    stage VARCHAR(50) NOT NULL,
    meta JSONB
    CONSTRAINT unique_timestamp UNIQUE (external_id, tag, stage)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS timestamps;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
