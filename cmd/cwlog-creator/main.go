package main

import (
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stax-labs/serverless-honeycomb-publisher/pkg/cwlogcreator"
)

func main() {
	lambda.Start(cfn.LambdaWrap(cwlogcreator.CreateLogGroupResource))
}
