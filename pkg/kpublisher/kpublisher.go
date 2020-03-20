package kpublisher

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/records"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// Publisher structure holding parsing things
type Publisher struct {
	parser parsers.LineParser
}

// New creates a Publisher
func New() (*Publisher, error) {
	parser, err := common.ConstructParser("json")
	if err != nil {
		return nil, err
	}

	return &Publisher{
		parser: parser,
	}, nil
}

// var parser parsers.LineParser
// var parserType, timeFieldName, timeFieldFormat, env string

// Handler takes a kinesis event and parses logs from it (sending them to honeycomb)
func (pb *Publisher) Handler(ctx context.Context, kinesisEvent events.KinesisEvent) (Response, error) {

	for _, record := range kinesisEvent.Records {
		kinesisRecord := record.Kinesis
		dataBytes := kinesisRecord.Data

		msgs, err := records.UncompressLogs(&kinesisRecord.ApproximateArrivalTimestamp.Time, dataBytes)
		if err != nil {
			logrus.WithError(err).Warn("unable to uncompress logs")
			continue
		}

		for _, msg := range msgs {

			if !strings.Contains(msg.Message, common.GetMatchString()) {
				continue
			}

			parsedLine, err := pb.parser.ParseLine(msg.Message)
			if err != nil {
				logrus.WithError(err).WithField("line", parsedLine).
					Warn("unable to parse line, skipping")
				continue
			}

			payload, err := common.ExtractPayload(parsedLine)
			if err != nil {
				logrus.WithError(err).WithField("line", parsedLine).
					Warn("unable to get event payload from line, skipping")
				continue
			}

			hnyEvent := libhoney.NewEvent()

			err = hnyEvent.Add(payload.Data)
			if err != nil {
				logrus.WithError(err).
					Warn("unable to add data to the event")
				continue
			}

			hnyEvent.AddField("aws.cloudwatch.loggroup", msg.LogGroup)

			// If we have sane values for other fields, set those as well
			if !payload.Time.IsZero() {
				hnyEvent.Timestamp = payload.Time
			}
			if payload.Dataset != "" {
				hnyEvent.Dataset = payload.Dataset
			}

			if payload.SampleRate > 0 {
				hnyEvent.SampleRate = payload.SampleRate
			}

			// We don't sample here - we assume it has been done upstream by
			// whatever wrote to the log
			err = hnyEvent.SendPresampled()
			if err != nil {
				logrus.WithError(err).
					Warn("unable to send the presampled event to honeycomb")
				continue
			}

		}

	}

	libhoney.Flush()

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}
