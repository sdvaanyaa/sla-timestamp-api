package postgres

import "errors"

var (
	ErrQueryFailed     = errors.New("database query execution error")
	ErrScanFailed      = errors.New("failed to extract row data")
	ErrRowsFailed      = errors.New("unexpected error during result iteration")
	ErrNoRowsAffected  = errors.New("no rows were affected by the operation")
	ErrMarshalFailed   = errors.New("failed to marshal meta data")
	ErrUnmarshalFailed = errors.New("failed to unmarshal meta data")
)
