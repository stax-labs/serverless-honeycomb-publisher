package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/honeycombio/libhoney-go"

	"github.com/sirupsen/logrus"

	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/kpublisher"
)

func main() {
	var err error

	logrus.Info("Starting Honeycomb Kinesis Publisher")

	if err = common.InitHoneycombFromEnvVars(); err != nil {
		logrus.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()

	common.AddUserAgentMetadata("publisher", "kinesis")

	kpub, err := kpublisher.New()
	if err != nil {
		logrus.WithError(err).
			Fatal("Unable to create a new Publisher")
	}

	lambda.Start(kpub.Handler)
}
