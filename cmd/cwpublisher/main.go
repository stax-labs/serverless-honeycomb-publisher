package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/honeycombio/honeytail/parsers"
	"github.com/honeycombio/libhoney-go"
	"github.com/sirupsen/logrus"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/common"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/cwpublisher"
)

var parser parsers.LineParser
var parserType, timeFieldName, timeFieldFormat, env string

func main() {
	var err error
	if err = common.InitHoneycombFromEnvVars(); err != nil {
		logrus.WithError(err).
			Fatal("Unable to initialize libhoney with the supplied environment variables")
		return
	}
	defer libhoney.Close()

	parser, err = common.ConstructParser("json")
	if err != nil {
		logrus.WithError(err).WithField("parser_type", parserType).
			Fatal("unable to construct parser")
		return
	}
	common.AddUserAgentMetadata("publisher", "json")

	lambda.Start(cwpublisher.Handler)
}
