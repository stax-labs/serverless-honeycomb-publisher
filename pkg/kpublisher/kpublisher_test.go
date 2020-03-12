package kpublisher

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
	"github.com/stretchr/testify/require"
)

func TestPublisher_Handler(t *testing.T) {
	assert := require.New(t)

	beeline.Init(beeline.Config{
		WriteKey:    "NOTTHEKEY",
		Dataset:     "TESTING",
		ServiceName: "SERVICE",
		STDOUT:      true,
	})

	parser, err := common.ConstructParser("json")
	assert.NoError(err)

	type fields struct {
		parser parsers.LineParser
	}
	type args struct {
		ctx          context.Context
		kinesisEvent events.KinesisEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name:   "expect ok no input values",
			fields: fields{parser: parser},
			args: args{
				context.TODO(), events.KinesisEvent{},
			},
			want: Response{Ok: true, Message: "ok"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := &Publisher{
				parser: tt.fields.parser,
			}
			got, err := pb.Handler(tt.args.ctx, tt.args.kinesisEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.Handler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(tt.want, got)
		})
	}
}
