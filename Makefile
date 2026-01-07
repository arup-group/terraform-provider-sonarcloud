# Package list and test args
TESTARGS?=
PKGS?=$$(go list ./... | grep -v 'vendor')

HOSTNAME=arup.com
NAMESPACE=platform
NAME=sonarcloud
VERSION?=0.1.0-local
OS_ARCH?=linux_amd64
BINARY_GLOB=./dist/terraform-provider-${NAME}_${OS_ARCH}/terraform*

ifneq (,$(wildcard .env))
    include .env
    export
endif

# Coverage and test flag configuration (centralized to avoid drift)
COVER?=
COVER_DIR=coverage
UNIT_COVER_PROFILE=$(COVER_DIR)/unit/coverage.out
ACC_COVER_PROFILE=$(COVER_DIR)/acceptance/coverage.out
BASE_UNIT_FLAGS?=-timeout=30s -parallel=4
BASE_ACC_FLAGS?=-timeout 120m -p 1 -v
COVER_FLAGS?=-covermode=atomic -coverpkg=./...

default: install

build:
	GORELEASER_CURRENT_TAG=$(VERSION) goreleaser build --snapshot --clean

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv $(BINARY_GLOB) ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	@if [ "$(COVER)" = "1" ]; then \
		mkdir -p $(COVER_DIR)/unit; \
		go test $(BASE_UNIT_FLAGS) $(TESTARGS) $(COVER_FLAGS) -coverprofile=$(UNIT_COVER_PROFILE) $(PKGS); \
	else \
		go test $(BASE_UNIT_FLAGS) $(TESTARGS) $(PKGS); \
	fi

# Convenience target for unit test coverage
test-coverage:
	COVER=1 $(MAKE) test

testacc:
	@if [ "$(COVER)" = "1" ]; then \
		mkdir -p $(COVER_DIR)/acceptance; \
		TF_ACC=1 go test $(BASE_ACC_FLAGS) $(TESTARGS) $(COVER_FLAGS) -coverprofile=$(ACC_COVER_PROFILE) $(PKGS); \
	else \
		TF_ACC=1 go test $(BASE_ACC_FLAGS) $(TESTARGS) $(PKGS); \
	fi

# Convenience target for acceptance test coverage
testacc-coverage:
	COVER=1 $(MAKE) testacc

debug-test:
	TF_ACC=true dlv test ./sonarcloud

fmt:
	go fmt ./sonarcloud

docs:
	go generate ./...

docs-check: docs
	git diff --quiet

.PHONY: docs
