package postgres

import "errors"

var (
	ErrQueryFailed     = errors.New("database query execution error")
	ErrScanFailed      = errors.New("failed to extract row data")
	ErrRowsFailed      = errors.New("unexpected error during result iteration")
	ErrUnmarshalFailed = errors.New("failed to unmarshal meta data")
)
