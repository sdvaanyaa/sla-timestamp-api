package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	smocks "github.com/sdvaanyaa/sla-timestamp-api/internal/repository/mocks"
	cmocks "github.com/sdvaanyaa/sla-timestamp-api/pkg/cache/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_timestampService_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()

	type fields struct {
		storageMock *smocks.TimestampStorageMock
		val         *validator.Validate
		cacheMock   *cmocks.CacheMock
	}
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name    string
		prepare func(ctx context.Context, a args, f *fields)
		args    args
		want    *entity.Timestamp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success Cache Hit",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				key := fmt.Sprintf(TimestampCachePrefix, a.id.String())
				ts := &entity.Timestamp{
					ID:         a.id,
					ExternalID: "test",
					Timestamp:  now,
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				}
				f.cacheMock.GetMock.Set(func(_ context.Context, k string, dest any) error {
					if k != key {
						return fmt.Errorf("unexpected key: %s", k)
					}
					if tsPtr, ok := dest.(*entity.Timestamp); ok {
						*tsPtr = *ts
						return nil
					}
					return fmt.Errorf("unexpected dest type: %T", dest)
				})
			},
			want: &entity.Timestamp{
				ID:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				ExternalID: "test",
				Timestamp:  now,
				Tag:        entity.TagIncident,
				Stage:      entity.StageCreated,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Success, Cache Miss, Storage Success",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				key := fmt.Sprintf(TimestampCachePrefix, a.id.String())
				f.cacheMock.GetMock.Expect(ctx, key, &entity.Timestamp{}).Return(errors.New("cache miss"))

				ts := &entity.Timestamp{
					ID:         a.id,
					ExternalID: "test",
					Timestamp:  now,
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				}
				f.storageMock.GetByIDMock.Expect(ctx, a.id).Return(ts, nil)

				f.cacheMock.SetMock.Expect(ctx, key, ts, CacheTTL).Return(nil)
			},
			want: &entity.Timestamp{
				ID:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				ExternalID: "test",
				Timestamp:  now,
				Tag:        entity.TagIncident,
				Stage:      entity.StageCreated,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Success, Cache Miss, Storage Success, Cache Set Error Ignored",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				key := fmt.Sprintf(TimestampCachePrefix, a.id.String())
				f.cacheMock.GetMock.Expect(ctx, key, &entity.Timestamp{}).Return(errors.New("cache miss"))

				ts := &entity.Timestamp{
					ID:         a.id,
					ExternalID: "test",
					Timestamp:  now,
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				}
				f.storageMock.GetByIDMock.Expect(ctx, a.id).Return(ts, nil)

				f.cacheMock.SetMock.Expect(ctx, key, ts, CacheTTL).Return(errors.New("set error"))
			},
			want: &entity.Timestamp{
				ID:         uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				ExternalID: "test",
				Timestamp:  now,
				Tag:        entity.TagIncident,
				Stage:      entity.StageCreated,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Input",
			args: args{
				id: uuid.Nil,
			},
			prepare: func(ctx context.Context, a args, f *fields) {
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Cache Miss, Storage Error",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				key := fmt.Sprintf(TimestampCachePrefix, a.id.String())
				f.cacheMock.GetMock.Expect(ctx, key, &entity.Timestamp{}).Return(errors.New("cache miss"))

				f.storageMock.GetByIDMock.Expect(ctx, a.id).Return(nil, errors.New("storage error"))
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			ctrl := minimock.NewController(t)
			storageMock := smocks.NewTimestampStorageMock(ctrl)
			cacheMock := cmocks.NewCacheMock(ctrl)

			s := &timestampService{
				storage: storageMock,
				val:     validator.New(),
				cache:   cacheMock,
			}

			tt.prepare(ctx, tt.args, &fields{
				storageMock: storageMock,
				cacheMock:   cacheMock,
				val:         validator.New(),
			})

			got, err := s.GetByID(ctx, tt.args.id)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
