BINDIR	:= bin/
ifeq ($(OS),Windows_NT)
	BINNAME	?= hypper.exe
else
	BINNAME	?= hypper
endif
INSTALL_PATH ?= /usr/local/bin

SHELL      = /usr/bin/env bash

GIT_COMMIT = $(shell git rev-parse HEAD)
ifneq ($(GIT_TAG),)
	GIT_TAG := $(GIT_TAG)
else
	GIT_TAG = $(shell git describe --tags 2>/dev/null)
endif

# go option
PKG        := ./...
TAGS       :=
TESTS      := .
TESTFLAGS  :=
LDFLAGS    := -w -s
GOFLAGS    :=

LDFLAGS += -X github.com/rancher-sandbox/hypper/internal/version.version=${GIT_TAG}
LDFLAGS += -X github.com/rancher-sandbox/hypper/internal/version.gitCommit=${GIT_COMMIT}
LDFLAGS += $(EXT_LDFLAGS)

.PHONY: all
all: build

.PHONY: build
build: lint $(BINDIR)$(BINNAME)

# Rebuild the binary if any of these files change
SRC := $(shell find . -type f -name '*.go' -print) go.mod go.sum

$(BINDIR)$(BINNAME): $(SRC)
	go build $(GOFLAGS) -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $(BINDIR)$(BINNAME) ./cmd/hypper

.PHONY: install
install: build
	@install "$(BINDIR)$(BINNAME)" "$(INSTALL_PATH)/$(BINNAME)"

.PHONY: test
test: lint build
test: TESTFLAGS += -race -v
test: test-style
test: test-unit

.PHONY: test-unit
test-unit:
	@echo "==> Running unit tests <=="
	go test $(GOFLAGS) -run $(TESTS) $(PKG) $(TESTFLAGS)

# Generate golden files used in unit tests
.PHONY: gen-test-golden
gen-test-golden:
gen-test-golden: PKG = ./cmd/hypper ./pkg/action
gen-test-golden: TESTFLAGS = -update
gen-test-golden: test-unit

.PHONY: test-style
test-style:
	@echo "==> Checking style <=="
	golangci-lint run

.PHONY: coverage
coverage:
	@echo "==> Running coverage tests <=="
	go test $(GOFLAGS) -run $(TESTS) $(PKG) -coverprofile=coverage.out --covermode=atomic

.PHONY: lint
lint: fmt vet

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: license-check
license-check:
	@scripts/license_check.sh

.PHONY: clean
clean:
	rm $(BINDIR)$(BINNAME)
	rmdir $(BINDIR)
