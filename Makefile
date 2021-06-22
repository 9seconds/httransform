ROOT_DIR   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

GOLANGCI_LINT_VERSION := v1.41.1

GOBIN  := $(ROOT_DIR)/.bin
GOTOOL := env "GOBIN=$(GOBIN)" "PATH=$(ROOT_DIR)/.bin:$(PATH)"

# -----------------------------------------------------------------------------

.PHONY: vendor
vendor: go.mod go.sum
	@go mod vendor

.bin:
	@mkdir -p "$(GOBIN)" || true

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

.PHONY: fmt
fmt:
	@$(GOTOOL) gofumpt -w -s -extra "$(ROOT_DIR)"

.PHONY: doc
doc:
	@$(GOTOOL) godoc -http 0.0.0.0:10000

.PHONY: install-tools
install-tools: install-tools-lint install-tools-godoc install-tools-gofumpt

.PHONY: install-tools-lint
install-tools-lint: .bin
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh \
		| bash -s -- -b "$(GOBIN)" "$(GOLANGCI_LINT_VERSION)"

.PHONY: install-tools-godoc
install-tools-godoc: .bin
	@$(GOTOOL) go get -u golang.org/x/tools/cmd/godoc

.PHONY: install-tools-gofumpt
install-tools-gofumpt: .bin
	@$(GOTOOL) go get -u mvdan.cc/gofumpt

.PHONY: upgrade-deps
upgrade-deps:
	@go get -u && go mod tidy
