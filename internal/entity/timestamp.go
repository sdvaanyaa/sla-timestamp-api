package entity

import (
	"github.com/google/uuid"
	"time"
)

type Timestamp struct {
	ID         uuid.UUID      `json:"id,omitempty"`
	ExternalID string         `json:"external_id" validate:"required"`
	Timestamp  time.Time      `json:"timestamp" validate:"required"`
	Tag        string         `json:"tag" validate:"required"`
	Stage      string         `json:"stage" validate:"required"`
	Meta       map[string]any `json:"meta,omitempty" validate:"omitempty"`
}

type CreateTimestampRequest struct {
	ExternalID string         `json:"external_id" validate:"required"`
	Timestamp  time.Time      `json:"timestamp" validate:"required"`
	Tag        string         `json:"tag" validate:"required"`
	Stage      string         `json:"stage" validate:"required"`
	Meta       map[string]any `json:"meta,omitempty" validate:"omitempty"`
}

func (r *CreateTimestampRequest) ToTimestamp() *Timestamp {
	return &Timestamp{
		ExternalID: r.ExternalID,
		Timestamp:  r.Timestamp,
		Tag:        r.Tag,
		Stage:      r.Stage,
		Meta:       r.Meta,
	}
}
