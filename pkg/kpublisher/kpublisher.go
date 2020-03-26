package kpublisher

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
	"github.com/pkg/errors"
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

		defer libhoney.Flush()

		for _, msg := range msgs {

			if !strings.Contains(msg.Message, common.GetMatchString()) {
				continue
			}

			hnyEvent, err := pb.buildEvent(msg.Message)
			if err != nil {
				logrus.WithError(err).Warn("unable to build event, skipping")
				continue
			}

			hnyEvent.AddField("aws.cloudwatch.loggroup", msg.LogGroup)

			// We don't sample here - we assume it has been done upstream by
			// whatever wrote to the log
			err = hnyEvent.SendPresampled()
			if err != nil {
				logrus.WithError(err).Warn("unable to send the presampled event to honeycomb")
				continue
			}

		}

	}

	logrus.WithField("count", len(kinesisEvent.Records)).Info("records processed")

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}

func (pb *Publisher) buildEvent(msg string) (*libhoney.Event, error) {
	parsedLine, err := pb.parser.ParseLine(msg)
	if err != nil {
		// logrus.WithError(err).WithField("line", parsedLine).
		// 	Warn("unable to parse line, skipping")
		return nil, errors.Wrap(err, "unable to parse line")
	}

	payload, err := common.ExtractPayload(parsedLine)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get event payload from line")
	}

	hnyEvent := libhoney.NewEvent()

	err = hnyEvent.Add(payload.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to add data to the event")
	}

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

	return hnyEvent, nil
}
