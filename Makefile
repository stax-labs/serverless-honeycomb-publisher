export STAGE
export BRANCH
export APP=honeycomb-publisher
export ROOT_DIR=$(pwd)
export HONEYCOMB_WRITE_KEY

PACKAGE_BUCKET ?= serverless-honeycomb-publisher-$(AWS_REGION)

test:
	@go test -v -cover ./publisher
.PHONY: test

clean:
	$(info [+] Cleanup dist folder")
	@rm -rf dist
.PHONY: clean	

validate-template:
	$(info [+] Validating cloudformation")
	aws cloudformation validate-template --template-body=file://sam/app/publisher.yml
.PHONY: validate-template

build:
	$(info [+] Build Lambda Binaries")
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/publisher ./publisher
	@GOOS=linux GOARCH=amd64 go build -o dist/cwlog-creator ./cwlog-creator
.PHONY: build

package:
	$(info [+] Package Binaries & Prepare Template")
	@cd dist && zip -X -9 -r ./handler.zip ./

	@aws cloudformation package \
		--template-file sam/app/publisher.yml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/publisher.out.yml
.PHONY: package

deployci:
	$(info [+] Deploy CI Pipeline")
	@aws cloudformation deploy \
		--template-file sam/ci/template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name serverless-agentless-honeycomb-publisher
		--parameter-overrides \
			GitHubOAuthTokenSecretId="github/pat"

.PHONY: deployci
