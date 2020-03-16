export STAGE
export BRANCH
export APP=honeycomb-publisher
export ROOT_DIR=$(pwd)
export HONEYCOMB_WRITE_KEY

GOLANGCI_VERSION = 1.23.6
PACKAGE_BUCKET ?= serverless-honeycomb-publisher-$(AWS_REGION)

test:
	@go test -v -cover ./pkg/cwpublisher
	@go test -v -cover ./pkg/kpublisher
	@go test -v -cover ./pkg/common
.PHONY: test

clean:
	$(info [+] Cleanup dist folder")
	@rm -rf dist
.PHONY: clean

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint 
	
bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | BINARY=golangci-lint bash -s -- v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

lint: bin/golangci-lint
	@echo "--- lint all the things"
	@bin/golangci-lint run
.PHONY: lint

validate-template:
	$(info [+] Validating cloudformation")
	aws cloudformation validate-template --template-body=file://sam/app/publisher.yml
.PHONY: validate-template

build:
	$(info [+] Build Lambda Binaries")
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -o dist/cwlog-creator ./cmd/cwlog-creator
	@GOOS=linux GOARCH=amd64 go build -o dist/cwpublisher ./cmd/cwpublisher
	@GOOS=linux GOARCH=amd64 go build -o dist/kpublisher ./cmd/kpublisher
.PHONY: build

packagezip:
	$(info [+] Package Binaries")
	@cd dist && zip -X -9 -r ./handler.zip ./

packagetest: packagezip
	$(info [+] Prepare Testing Template")
	aws cloudformation package \
		--template-file sam/testing/template.yml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/test-packaged-template.yml
.PHONY: packagetest

deploytest:
	$(info [+] Deploy Testing Publishers")
	@aws cloudformation deploy \
		--template-file dist/test-packaged-template.yml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP)-$(STAGE)-$(BRANCH) \
		--parameter-overrides Stage=$(STAGE)
.PHONY: deploytest

package: packagezip
	$(info [+] Prepare Template")
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
