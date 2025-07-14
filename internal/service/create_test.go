package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	"github.com/sdvaanyaa/sla-timestamp-api/internal/entity"
	smocks "github.com/sdvaanyaa/sla-timestamp-api/internal/repository/mocks"
	bmocks "github.com/sdvaanyaa/sla-timestamp-api/pkg/broker/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_timestampService_Create(t *testing.T) {
	t.Parallel()

	type fields struct {
		storageMock *smocks.TimestampStorageMock
		val         *validator.Validate
		brokerMock  *bmocks.BrokerMock
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

				a.ts.ID = id
				event := map[string]any{"action": "create", "data": a.ts}
				msg, _ := json.Marshal(event)
				f.brokerMock.PublishMock.Expect(ctx, msg).Return(nil)
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
			name: "Publish Error (Ignored)",
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

				a.ts.ID = id
				event := map[string]any{"action": "create", "data": a.ts}
				msg, _ := json.Marshal(event)
				f.brokerMock.PublishMock.Expect(ctx, msg).Return(errors.New("publish error"))
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
			brokerMock := bmocks.NewBrokerMock(ctrl)

			s := &timestampService{
				storage: storageMock,
				val:     validator.New(),
				broker:  brokerMock,
			}

			tt.prepare(ctx, tt.args, &fields{
				storageMock: storageMock,
				val:         validator.New(),
				brokerMock:  brokerMock,
			})

			got, err := s.Create(ctx, tt.args.ts)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
