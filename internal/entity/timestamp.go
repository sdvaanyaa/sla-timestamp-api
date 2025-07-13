package entity

import (
	"github.com/google/uuid"
	"time"
)

type Tag string
type Stage string

const (
	TagIncident    Tag = "incident"
	TagSLA         Tag = "sla"
	TagDeployment  Tag = "deployment"
	TagMaintenance Tag = "maintenance"
	TagAlert       Tag = "alert"
)

const (
	StageCreated      Stage = "created"
	StageAcknowledged Stage = "acknowledged"
	StageInProgress   Stage = "in_progress"
	StageResolved     Stage = "resolved"
	StageClosed       Stage = "closed"
)

type Timestamp struct {
	ID         uuid.UUID      `json:"id,omitempty"`
	ExternalID string         `json:"external_id" validate:"required"`
	Timestamp  time.Time      `json:"timestamp" validate:"required"`
	Tag        Tag            `json:"tag" validate:"required,oneof=incident sla deployment maintenance alert"`
	Stage      Stage          `json:"stage" validate:"required,oneof=created acknowledged in_progress resolved closed"`
	Meta       map[string]any `json:"meta,omitempty" validate:"omitempty"`
}

type CreateTimestampRequest struct {
	ExternalID string         `json:"external_id" validate:"required"`
	Timestamp  time.Time      `json:"timestamp" validate:"required" example:"2025-07-13T15:00:00Z"`
	Tag        Tag            `json:"tag" validate:"required,oneof=incident sla deployment maintenance alert"`
	Stage      Stage          `json:"stage" validate:"required,oneof=created acknowledged in_progress resolved closed"`
	Meta       map[string]any `json:"meta,omitempty" validate:"omitempty"`
}

type ListQueryParams struct {
	Limit         int    `validate:"gte=1"`
	Offset        int    `validate:"gte=0"`
	ExternalID    string `validate:"omitempty"`
	Tag           string `validate:"omitempty,oneof=incident sla deployment maintenance alert"`
	Stage         string `validate:"omitempty,oneof=created acknowledged in_progress resolved closed"`
	TimestampFrom *time.Time
	TimestampTo   *time.Time
	MetaFilter    map[string]any `validate:"omitempty"`
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
