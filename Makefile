# TOOLS VERSIONS
GO_VERSION=1.21.0
GOLANGCI_LINT_VERSION=v1.54.0

image=perebaj
version=$(shell git rev-parse --short HEAD)

## run isolated tests
.PHONY: test
test:
	go test ./... -timeout 10s -race

## builds the service
.PHONY: service
service:
	go build -o ./cmd/grime/grime ./cmd/grime

## runs the service locally
.PHONY: run
run: service
	./cmd/grime/grime

## lint the whole project
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...


## Build the service image
.PHONY: image
image:
	docker build . \
		--build-arg GO_VERSION=$(GO_VERSION) \
		-t $(image)

## Publish the service image
.PHONY: image/publish
image/publish: image
	docker push $(image)

## Display help for all targets
.PHONY: help
help:
	@awk '/^.PHONY: / { \
		msg = match(lastLine, /^## /); \
			if (msg) { \
				cmd = substr($$0, 9, 100); \
				msg = substr(lastLine, 4, 1000); \
				printf "  ${GREEN}%-30s${RESET} %s\n", cmd, msg; \
			} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
