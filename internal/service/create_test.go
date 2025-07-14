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

func Test_timestampService_Create(t *testing.T) {
	t.Parallel()

	type fields struct {
		storageMock *smocks.TimestampStorageMock
		val         *validator.Validate
		cacheMock   *cmocks.CacheMock
	}
	type args struct {
		ts *entity.Timestamp
	}
	tests := []struct {
		name    string
		prepare func(ctx context.Context, a args, f *fields)
		args    args
		want    uuid.UUID
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			args: args{
				ts: &entity.Timestamp{
					ExternalID: "test",
					Timestamp:  time.Now(),
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				},
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				f.storageMock.CreateMock.Expect(ctx, a.ts).Return(id, nil)

				key := fmt.Sprintf(TimestampCachePrefix, id.String())
				f.cacheMock.SetMock.Expect(ctx, key, a.ts, CacheTTL).Return(nil)
				f.cacheMock.DeleteMock.Expect(ctx, ListCachePrefix).Return(nil)
			},
			want:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Input",
			args: args{
				ts: &entity.Timestamp{
					ExternalID: "", //  required
					Timestamp:  time.Now(),
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				},
			},
			prepare: func(ctx context.Context, a args, f *fields) {
			},
			want:    uuid.Nil,
			wantErr: assert.Error,
		},
		{
			name: "Storage Error",
			args: args{
				ts: &entity.Timestamp{
					ExternalID: "test",
					Timestamp:  time.Now(),
					Tag:        entity.TagAlert,
					Stage:      entity.StageCreated,
				},
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				f.storageMock.CreateMock.Expect(ctx, a.ts).Return(uuid.Nil, errors.New("storage error"))
			},
			want:    uuid.Nil,
			wantErr: assert.Error,
		},
		{
			name: "Cache Set Error (Ignored)",
			args: args{
				ts: &entity.Timestamp{
					ExternalID: "test",
					Timestamp:  time.Now(),
					Tag:        entity.TagDeployment,
					Stage:      entity.StageCreated,
				},
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

				f.storageMock.CreateMock.Expect(ctx, a.ts).Return(id, nil)

				key := fmt.Sprintf(TimestampCachePrefix, id.String())
				f.cacheMock.SetMock.Expect(ctx, key, a.ts, CacheTTL).Return(errors.New("cache set error"))
				f.cacheMock.DeleteMock.Expect(ctx, ListCachePrefix).Return(nil)
			},
			want:    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			wantErr: assert.NoError,
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

			got, err := s.Create(ctx, tt.args.ts)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
