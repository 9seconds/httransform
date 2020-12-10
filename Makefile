ROOT_DIR   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

GOLANGCI_LINT_VERSION := v1.33.0

GOBIN  := $(ROOT_DIR)/.bin
GOTOOL := env "GOBIN=$(GOBIN)" "PATH=$(ROOT_DIR)/.bin:$(PATH)"

# -----------------------------------------------------------------------------

.PHONY: all
all: build

.PHONY: build
	@go build

.PHONY: test
test:
	@go test -v ./...

.PHONY: citest
citest:
	@go test  -coverprofile=coverage.txt -covermode=atomic -race -v ./...

.PHONY: clean
clean:
	@git clean -xfd && \
		git reset --hard >/dev/null && \
		git submodule foreach --recursive sh -c 'git clean -xfd && git reset --hard' >/dev/null

.PHONY: lint
lint:
	@$(GOTOOL) golangci-lint run

.PHONY: doc
doc:
	@$(GOTOOL) godoc -http 0.0.0.0:10000

.PHONY: install-tools
install-tools: install-tools-lint install-tools-godoc

.PHONY: install-tools-lint
install-tools-lint:
	@mkdir -p "$(GOBIN)" || true && \
		curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh \
		| bash -s -- -b "$(GOBIN)" "$(GOLANGCI_LINT_VERSION)"

.PHONY: install-tools-godoc
install-tools-godoc:
	@mkdir -p "$(GOBIN)" || true && \
		$(GOTOOL) go get -u golang.org/x/tools/cmd/godoc

.PHONY: update-deps
upgrade-deps:
	$go get -u && go mod tidy
