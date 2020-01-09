package main

import (
	"context"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/pkg/errors"
)

// createLogGroupResource this lambda will try to create a log group and if it exists just pass through success
//
// This function requires a LogGroupName to be present in the properties passed in the event.
//
func createLogGroupResource(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {

	logGroupName, ok := event.ResourceProperties["LogGroupName"].(string)
	if !ok {
		return "", nil, errors.New("missing required LogGroupName from Properties")
	}

	data := map[string]interface{}{"LogGroupName": logGroupName}

	sess := session.Must(session.NewSession())
	cwlogsvc := cloudwatchlogs.New(sess)

	switch event.RequestType {
	case cfn.RequestCreate, cfn.RequestUpdate: // only run on create / update
		_, err := cwlogsvc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
			LogGroupName: aws.String(logGroupName),
		})
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == cloudwatchlogs.ErrCodeResourceAlreadyExistsException {
				// as the resource already exists we can just return the log group name
				return "", data, nil
			}
			return "", nil, errors.Wrap(err, "failed to create log group")
		}
	}

	return "", data, nil
}

func main() {
	lambda.Start(cfn.LambdaWrap(createLogGroupResource))
}
