GO ?= go
GOFMT ?= gofmt "-s"
# GO_VERSION=$(shell $(GO) version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f2)
PACKAGES ?= $(shell $(GO) list ./...)
VETPACKAGES ?= $(shell $(GO) list ./...)
GOFILES := $(shell find . -name "*.go")
TESTFOLDER := $(shell $(GO) list ./...)


.PHONY: test
# Run tests to verify code functionality.
test:
	$(GO) test -v $(TESTFOLDER)
# Run tests with data race detector.
.PHONY: race
race:
	$(GO) test -race -v $(TESTFOLDER)
.PHONY: benchmark
# Run benchmarks to evaluate code performance.
benchmark:
	$(GO) test -bench=".*" $(TESTFOLDER)

.PHONY: fmt
# Ensure consistent code formatting.
fmt:
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
# format (check only).
fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: vet
# Examine packages and report suspicious constructs if any.
vet:
	$(GO) vet $(VETPACKAGES)

.PHONY: lint
# Inspect source code for stylistic errors or potential bugs.
lint:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

.PHONY: tools
# Install tools
tools:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

.PHONY: help
# Help.
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf " - \033[36m%-20s\033[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help