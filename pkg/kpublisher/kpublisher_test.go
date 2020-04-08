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

func TestPublisher_buildEvent(t *testing.T) {

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
		msg string
	}
	type want struct {
		dataset string
		spanID  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name:   "expect ok no input values",
			fields: fields{parser: parser},
			args: args{
				msg: `{"samplerate":1,"dataset":"travis.serverless-test","data":{"trace.span_id":"28320411-35c9-45dc-b968-b8b9d091f4bf","meta.local_hostname":"ip-10-12-92-115","trace.trace_id":"6614adb4-a74e-47a2-92de-06435f548eb1","name":"sleepytime","service_name":"travis.serverless-test","trace.parent_id":"56c10d68-e8db-46d5-97ad-0d508a3c08bf","duration_ms":500.65700000000004,"meta.beeline_version":"1.0.0"},"user_agent":"libhoney-py/1.4.0","time":"2018-07-23T22:06:58.471593Z"}`,
			},
			want: want{
				dataset: "travis.serverless-test",
				spanID:  "28320411-35c9-45dc-b968-b8b9d091f4bf",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := &Publisher{
				parser: tt.fields.parser,
			}
			got, err := pb.buildEvent(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Publisher.buildEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(tt.want.dataset, got.Dataset)
			assert.Equal(tt.want.spanID, got.Fields()["trace.span_id"])
		})
	}
}
