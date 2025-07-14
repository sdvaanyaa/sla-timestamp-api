package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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

func Test_timestampService_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	type fields struct {
		storageMock *smocks.TimestampStorageMock
		val         *validator.Validate
		cacheMock   *cmocks.CacheMock
	}
	type args struct {
		limit         int
		offset        int
		externalID    string
		tag           string
		stage         string
		timestampFrom *time.Time
		timestampTo   *time.Time
		metaFilter    map[string]any
	}
	tests := []struct {
		name    string
		prepare func(ctx context.Context, a args, f *fields)
		args    args
		want    []*entity.Timestamp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success, Cache Miss, Storage Success",
			args: args{
				limit:      10,
				offset:     0,
				externalID: "test",
				tag:        "incident",
				stage:      "created",
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				params := &entity.ListQueryParams{
					Limit:      a.limit,
					Offset:     a.offset,
					ExternalID: a.externalID,
					Tag:        a.tag,
					Stage:      a.stage,
				}
				paramsJSON, _ := json.Marshal(params)
				key := fmt.Sprintf("timestamps:list:%x", sha256.Sum256(paramsJSON))

				var dest []*entity.Timestamp
				f.cacheMock.GetMock.Expect(ctx, key, &dest).Return(errors.New("cache miss"))

				list := []*entity.Timestamp{
					{
						ID:         id,
						ExternalID: "test",
						Timestamp:  now,
						Tag:        entity.TagIncident,
						Stage:      entity.StageCreated,
					},
				}
				f.storageMock.ListMock.Expect(ctx, a.limit, a.offset, a.externalID, a.tag, a.stage, a.timestampFrom, a.timestampTo, a.metaFilter).Return(list, nil)

				f.cacheMock.SetMock.Expect(ctx, key, list, CacheTTL).Return(nil)
			},
			want: []*entity.Timestamp{
				{
					ID:         id,
					ExternalID: "test",
					Timestamp:  now,
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Success, Cache Miss, Storage Success, Cache Set Error Ignored",
			args: args{
				limit:      10,
				offset:     0,
				externalID: "test",
				tag:        "incident",
				stage:      "created",
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				params := &entity.ListQueryParams{
					Limit:      a.limit,
					Offset:     a.offset,
					ExternalID: a.externalID,
					Tag:        a.tag,
					Stage:      a.stage,
				}
				paramsJSON, _ := json.Marshal(params)
				key := fmt.Sprintf("timestamps:list:%x", sha256.Sum256(paramsJSON))

				var dest []*entity.Timestamp
				f.cacheMock.GetMock.Expect(ctx, key, &dest).Return(errors.New("cache miss"))

				list := []*entity.Timestamp{
					{
						ID:         id,
						ExternalID: "test",
						Timestamp:  now,
						Tag:        entity.TagIncident,
						Stage:      entity.StageCreated,
					},
				}
				f.storageMock.ListMock.Expect(
					ctx,
					a.limit,
					a.offset,
					a.externalID,
					a.tag,
					a.stage,
					a.timestampFrom,
					a.timestampTo,
					a.metaFilter,
				).Return(list, nil)

				f.cacheMock.SetMock.Expect(ctx, key, list, CacheTTL).Return(errors.New("set error"))
			},
			want: []*entity.Timestamp{
				{
					ID:         id,
					ExternalID: "test",
					Timestamp:  now,
					Tag:        entity.TagIncident,
					Stage:      entity.StageCreated,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Params",
			args: args{
				limit: 0,
			},
			prepare: func(ctx context.Context, a args, f *fields) {
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "Cache Miss, Storage Error",
			args: args{
				limit:      10,
				offset:     0,
				externalID: "test",
				tag:        "incident",
				stage:      "created",
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				params := &entity.ListQueryParams{
					Limit:      a.limit,
					Offset:     a.offset,
					ExternalID: a.externalID,
					Tag:        a.tag,
					Stage:      a.stage,
				}
				paramsJSON, _ := json.Marshal(params)
				key := fmt.Sprintf("timestamps:list:%x", sha256.Sum256(paramsJSON))

				var dest []*entity.Timestamp
				f.cacheMock.GetMock.Expect(ctx, key, &dest).Return(errors.New("cache miss"))

				f.storageMock.ListMock.Expect(
					ctx,
					a.limit,
					a.offset,
					a.externalID,
					a.tag,
					a.stage,
					a.timestampFrom,
					a.timestampTo,
					a.metaFilter,
				).Return(nil, errors.New("storage error"))
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

			got, err := s.List(
				ctx,
				tt.args.limit,
				tt.args.offset,
				tt.args.externalID,
				tt.args.tag,
				tt.args.stage,
				tt.args.timestampFrom,
				tt.args.timestampTo,
				tt.args.metaFilter,
			)

			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_timestampService_validateListParams(t *testing.T) {
	t.Parallel()

	type fields struct {
		val *validator.Validate
	}
	type args struct {
		params *entity.ListQueryParams
	}

	now := time.Now()
	from := now.Add(24 * time.Hour)
	to := now

	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success Valid Params",
			args: args{
				params: &entity.ListQueryParams{
					Limit:  10,
					Offset: 0,
					Tag:    "incident",
					Stage:  "created",
					MetaFilter: map[string]any{
						"source": "email",
					},
				},
			},
			prepare: func(f *fields) {},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Validator",
			args: args{
				params: &entity.ListQueryParams{
					Limit: 0, // gte=1
				},
			},
			prepare: func(f *fields) {},
			wantErr: assert.Error,
		},
		{
			name: "Invalid Empty Meta Key",
			args: args{
				params: &entity.ListQueryParams{
					Limit:  10,
					Offset: 0,
					MetaFilter: map[string]any{
						"": "value",
					},
				},
			},
			prepare: func(f *fields) {},
			wantErr: assert.Error,
		},
		{
			name: "Invalid Timestamp From After To",
			args: args{
				params: &entity.ListQueryParams{
					Limit:         10,
					Offset:        0,
					TimestampFrom: &from,
					TimestampTo:   &to,
				},
			},
			prepare: func(f *fields) {},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := &fields{
				val: validator.New(),
			}
			tt.prepare(f)

			s := timestampService{
				val: f.val,
			}

			err := s.validateListParams(tt.args.params)
			tt.wantErr(t, err)
		})
	}
}
