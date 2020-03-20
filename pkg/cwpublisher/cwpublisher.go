package cwpublisher

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
)

// Response is a simple structured response
type Response struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// Handler takes a cloudwatch event and parses logs from it (sending them to honeycomb)
func Handler(request events.CloudwatchLogsEvent) (Response, error) {

	parser, err := common.ConstructParser("json")
	if err != nil {
		return Response{
			Ok:      false,
			Message: "parser not initialized, cannot process events",
		}, fmt.Errorf("parser not initialized, cannot process events")
	}
	common.AddUserAgentMetadata("publisher", "json")

	data, err := request.AWSLogs.Parse()
	if err != nil {
		return Response{
			Ok:      false,
			Message: fmt.Sprintf("failed to parse cloudwatch event data: %s", err.Error()),
		}, err
	}

	for _, event := range data.LogEvents {
		parsedLine, err := parser.ParseLine(event.Message)
		if err != nil {
			logrus.WithError(err).WithField("line", event.Message).
				Warn("unable to parse line, skipping")
			continue
		}
		// The JSON parser returns a map[string]interface{} - we need to convert it
		// to a structure we can work with
		payload, err := common.ExtractPayload(parsedLine)
		if err != nil {
			logrus.WithError(err).WithField("line", event.Message).
				Warn("unable to get event payload from line, skipping")
			continue
		}
		hnyEvent := libhoney.NewEvent()
		// add the actual event data
		err = hnyEvent.Add(payload.Data)
		if err != nil {
			logrus.WithError(err).
				Warn("unable to add data to the event")
			continue
		}

		// Include the logstream that this data came from to make it easier to find the source
		// in Cloudwatch
		hnyEvent.AddField("aws.cloudwatch.logstream", data.LogStream)

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

	libhoney.Flush()

	return Response{
		Ok:      true,
		Message: "ok",
	}, nil
}
