ROOT_DIR   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

GOLANGCI_LINT_VERSION := v1.13.2

MOD_ON  := env GO111MODULE=on
MOD_OFF := env GO111MODULE=auto

# -----------------------------------------------------------------------------

vendor: go.mod go.sum
	@$(MOD_ON) go mod vendor

# -----------------------------------------------------------------------------

.PHONY: all
all:
	@$(MOD_ON) go build

.PHONY: test
test:
	@$(MOD_ON) go test -v ./...

.PHONY: citest
citest:
	@$(MOD_ON) go test  -coverprofile=coverage.txt -covermode=atomic -race -v ./...

.PHONY: lint
lint:
	@$(MOD_ON) golangci-lint run

.PHONY: clean
clean:
	@git clean -xfd && \
		git reset --hard >/dev/null && \
		git submodule foreach --recursive sh -c 'git clean -xfd && git reset --hard' >/dev/null

.PHONY: prepare
prepare: install-lint

.PHONY: install-lint
install-lint:
	@curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh \
		| $(MOD_OFF) bash -s -- -b $(GOPATH)/bin $(GOLANGCI_LINT_VERSION)
