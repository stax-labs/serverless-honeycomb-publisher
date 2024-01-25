export STAGE
export BRANCH
export APP=honeycomb-publisher
export ROOT_DIR=$(pwd)
export HONEYCOMB_WRITE_KEY

PACKAGE_BUCKET 		?= serverless-honeycomb-publisher-$(AWS_REGION)
GIT_HASH    		?= $(shell git rev-parse --short HEAD)

# Go related vars
# ----------------------
GOLANGCI_VERSION 	:= 1.55.2

BUILD_OVERRIDES = \
	-X "$(PACKAGE)/internal/app.Name=$(APP)" \
	-X "$(PACKAGE)/internal/app.BuildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')" \
	-X "$(PACKAGE)/internal/app.Commit=$(GIT_HASH)" \
# the -w -s flags make the binary a bit smaller and
# trimpath shortens build paths in stack traces
LDFLAGS := -ldflags='-w -s $(BUILD_OVERRIDES)' -trimpath
export GOFLAGS=-buildvcs=false

test:
	@go test -v -cover ./...
.PHONY: test

clean:
	$(info [+] Cleanup dist folder)
	@rm -rf dist
.PHONY: clean

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint

bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v$(GOLANGCI_VERSION)
	@mv bin/golangci-lint $@

lint: bin/golangci-lint
	$(info [+] Linting)
	@bin/golangci-lint run --timeout 300s ./...
.PHONY: lint

validate-template:
	$(info [+] Validating cloudformation)
	aws cloudformation validate-template --template-body=file://sam/app/publisher.yml
.PHONY: validate-template

build:
	$(info [+] Build Lambda Binaries)
	@mkdir -p dist
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a $(LDFLAGS) -o dist/cwlog-creator/bootstrap ./cmd/cwlog-creator
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a $(LDFLAGS) -o dist/cwpublisher/bootstrap ./cmd/cwpublisher
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -a $(LDFLAGS) -o dist/kpublisher/bootstrap ./cmd/kpublisher
.PHONY: build

packagezip:
	$(info [+] Package Binaries)
	@zip -j -q dist/cwlog-creator.zip dist/cwlog-creator/bootstrap
	@zip -j -q dist/cwpublisher.zip dist/cwpublisher/bootstrap
	@zip -j -q dist/kpublisher.zip dist/kpublisher/bootstrap

packagetest: packagezip
	$(info [+] Prepare Testing Template)
	aws cloudformation package \
		--template-file sam/testing/template.yml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/test-packaged-template.yml
.PHONY: packagetest

deploytest:
	$(info [+] Deploy Testing Publishers)
	@aws cloudformation deploy \
		--template-file dist/test-packaged-template.yml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name $(APP)-$(STAGE)-$(BRANCH) \
		--parameter-overrides Stage=$(STAGE)
.PHONY: deploytest

package: packagezip
	$(info [+] Prepare Template)
	@aws cloudformation package \
		--template-file sam/app/publisher.yml \
		--s3-bucket $(PACKAGE_BUCKET) \
		--output-template-file dist/publisher.out.yml
.PHONY: package

deployci:
	$(info [+] Deploy CI Pipeline)
	@aws cloudformation deploy \
		--template-file sam/ci/template.yaml \
		--capabilities CAPABILITY_NAMED_IAM CAPABILITY_IAM CAPABILITY_AUTO_EXPAND \
		--stack-name serverless-agentless-honeycomb-publisher
		--parameter-overrides \
			GitHubOAuthTokenSecretId="github/pat"

.PHONY: deployci
