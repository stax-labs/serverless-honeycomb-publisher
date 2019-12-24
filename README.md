# Serverless Honeycomb Publisher

This repo extends the agentless publisher provided by [Honeycomb](https://www.honeycomb.io/) in their [agentless-integrations-for-aws](https://github.com/honeycombio/agentless-integrations-for-aws) by uploading the go binary and SAM template to Amazons AWS Serverless Application Repository for use in CloudFormation Templates.

If you would like to use this application in your CloudFormation see the [Usage](#usage) section of this README for details on how.

This Serverless Application is publicly available at the ARN: 
* `arn:aws:serverlessrepo:us-east-1:541595141780:applications/serverless-agentless-honeycomb-publisher`

If you would like to fork and host the serverless honeycomb publisher privately this repo also contains the AWS Code Pipeline definition necessary for this, the details are in the [Deployment](#deployment) guide and make use of the AWS Labs provided [aws-sam-codepipeline-cd](https://github.com/awslabs/aws-sam-codepipeline-cd)

## Purpose (Why?)

The agentless-honeycomb-publisher works really well for our use-case, allowing us to write logs to STDOUT then pick them up asyncronously from CloudWatch Logs, however Stax runs lambda functions in many AWS Accounts and AWS Regions, so rather than deal with creating even more pipelines and code to roll out publishers into each environment, when all we really want is the publisher itself we figured why not host it in the AWS Serverless Repo and deploy "once".

[Serverless Agentless Honeycomb Publisher](https://serverlessrepo.aws.amazon.com/applications/arn:aws:serverlessrepo:us-east-1:541595141780:applications~serverless-agentless-honeycomb-publisher)

Then we figured if we were going to maintain this for ourselves, why not share it so others can either use the publisher from the SAR directly, or build their own pipeline.

![Simple Publishing Diagram](https://github.com/stax-labs/serverless-honeycomb-publisher/raw/master/images/simple-publishing-diagram.png)

This also allows us to extend upon the functionality in the Publisher.

## Usage

You can use this from any region or AWS Account using the follow resource in your template:

**yaml**
```yaml
HoneycombPublisher:
    Type: 'AWS::Serverless::Application'
    Properties:
    Location:
        ApplicationId: arn:aws:serverlessrepo:us-east-1:541595141780:applications/serverless-agentless-honeycomb-publisher
        SemanticVersion: 1.0.0
    Parameters:
        HoneycombWriteKey: <YOURKEY>
        HoneycombDataset: <YOURDATASET>
        LogGroupName: <YOURLOGGROUP>
        FilterPattern: ""
```

**json**
```json
{
    "HoneycombPublisher": {
    "Type": "AWS::Serverless::Application",
        "Properties": {
            "Location": {
                "ApplicationId": "arn:aws:serverlessrepo:us-east-1:541595141780:applications/serverless-agentless-honeycomb-publisher",
                "SemanticVersion": "1.0.0"
            },
            "Parameters": {
                "HoneycombDataset": "<YOURDATASET>",
                "HoneycombWriteKey": "<YOURWRITEKEY>",
                "FilterPattern": "",
                "LogGroupName": "<YOURLOGGROUP>",
            }
        }
    }
}
```

Optionally you can specify up to five additional log groups:
* LogGroupName1
* LogGroupName2
* LogGroupName3
* LogGroupName4
* LogGroupName5

As well as the KMS KeyID for decrpytion of your token:
* KMSKeyId

And the Honeycomb API Host:
* HoneycombAPIHost

## Deployment

These steps are only required if you would like to host the application in your own AWS Account, if you would like to make use of the already deployed version see the [Usage](#usage) section of this document.

### Requirements

* AWS Account (Application must be deployed in us-east-1 if you intend to share it publicly)
* A fork of this repository in Github (other SCM tools are supported by [aws-sam-codepipeline-cd](https://github.com/awslabs/aws-sam-codepipeline-cd) but are beyond the scope of this readme)
* A PAT token used by AWS CodeBuild to connect to Github

### Steps

1. Update the Makefile to specify your own:
    * S3 Bucket
    * GitHubOAuthTokenSecretId (in our case this is the path to a secret in AWS Secrets Manager)
1. Run the make target deployci; you will need an active AWS session with the appropriate permissions for this deployment.

    ```bash
    $ make deployci
    ```

## License

This application was released under the Apache 2.0 license.

## Sponsor

This project is sponsored by [Stax](https://stax.io), a dedicated platform to accelerate your cloud journey.

![Stax Logo](https://github.com/stax-labs/serverless-honeycomb-publisher/raw/master/images/stax-logo.png)