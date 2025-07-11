package entity

import (
	"github.com/google/uuid"
	"time"
)

type Timestamp struct {
	ID         uuid.UUID      `json:"id,omitempty"`
	ExternalID string         `json:"external_id"`
	Timestamp  time.Time      `json:"timestamp"`
	Tag        string         `json:"tag"`
	Stage      string         `json:"stage"`
	Meta       map[string]any `json:"meta,omitempty"`
}
