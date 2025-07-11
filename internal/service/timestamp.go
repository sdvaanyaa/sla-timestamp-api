package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/repository"
)

var ErrInvalidInput = errors.New("invalid input")

type TimestampService interface {
	Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Timestamp, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type timestampService struct {
	storage repository.TimestampStorage
}

func New(storage repository.TimestampStorage) TimestampService {
	return &timestampService{
		storage: storage,
	}
}

func (s *timestampService) Create(ctx context.Context, ts *entity.Timestamp) (uuid.UUID, error) {
	if ts.ExternalID == "" || ts.Tag == "" || ts.Stage == "" {
		return uuid.Nil, ErrInvalidInput
	}

	return s.storage.Create(ctx, ts)
}

func (s *timestampService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Timestamp, error) {
	if id == uuid.Nil {
		return nil, ErrInvalidInput
	}

	return s.storage.GetByID(ctx, id)
}

func (s *timestampService) List(ctx context.Context, limit, offset int) ([]*entity.Timestamp, error) {
	if limit <= 0 || offset < 0 {
		return nil, ErrInvalidInput
	}

	return s.storage.List(ctx, limit, offset)
}

func (s *timestampService) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrInvalidInput
	}

	return s.storage.Delete(ctx, id)
}
