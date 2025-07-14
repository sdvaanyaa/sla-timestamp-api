package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/gojuno/minimock/v3"
	"github.com/google/uuid"
	smocks "github.com/sdvaanyaa/sla-timestamp-api/internal/repository/mocks"
	bmocks "github.com/sdvaanyaa/sla-timestamp-api/pkg/broker/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_timestampService_Delete(t *testing.T) {
	t.Parallel()

	type fields struct {
		storageMock *smocks.TimestampStorageMock
		val         *validator.Validate
		brokerMock  *bmocks.BrokerMock
	}
	type args struct {
		id uuid.UUID
	}
	tests := []struct {
		name    string
		prepare func(ctx context.Context, a args, f *fields)
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				f.storageMock.DeleteMock.Expect(ctx, a.id).Return(nil)

				event := map[string]any{"action": "delete", "id": a.id.String()}
				msg, _ := json.Marshal(event)
				f.brokerMock.PublishMock.Expect(ctx, msg).Return(nil)
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
			wantErr: assert.Error,
		},
		{
			name: "Storage Error",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				f.storageMock.DeleteMock.Expect(ctx, a.id).Return(errors.New("storage error"))
			},
			wantErr: assert.Error,
		},
		{
			name: "Cache Delete Error (Ignored)",
			args: args{
				id: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			},
			prepare: func(ctx context.Context, a args, f *fields) {
				f.storageMock.DeleteMock.Expect(ctx, a.id).Return(nil)

				event := map[string]any{"action": "delete", "id": a.id.String()}
				msg, _ := json.Marshal(event)
				f.brokerMock.PublishMock.Expect(ctx, msg).Return(errors.New("publish error"))
			},
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

			s := timestampService{
				storage: storageMock,
				val:     validator.New(),
				broker:  brokerMock,
			}

			tt.prepare(ctx, tt.args, &fields{
				storageMock: storageMock,
				brokerMock:  brokerMock,
				val:         validator.New(),
			})

			err := s.Delete(ctx, tt.args.id)
			tt.wantErr(t, err)
		})
	}
}
