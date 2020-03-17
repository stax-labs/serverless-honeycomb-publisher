package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/cwpublisher"
)

func main() {
	var err error
	if err = common.InitHoneycombFromEnvVars(); err != nil {
		logrus.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()

	lambda.Start(cwpublisher.Handler)
}
