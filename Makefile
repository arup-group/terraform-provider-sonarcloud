# Package list and test args
TESTARGS?=
PKGS?=$$(go list ./... | grep -v 'vendor')

HOSTNAME=arup.com
NAMESPACE=platform
NAME=sonarcloud
VERSION?=0.1.0-local
OS_ARCH?=linux_amd64
# Use wildcard to handle goreleaser's version suffix (e.g., linux_amd64_v1)
BINARY_GLOB=./dist/terraform-provider-${NAME}_${OS_ARCH}*/terraform-provider-*

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
	set -- $(BINARY_GLOB); \
	if [ "$$#" -ne 1 ] || [ ! -f "$$1" ]; then \
		echo "Expected exactly one build artifact matching '$(BINARY_GLOB)', but found $$# or no valid file."; \
		exit 1; \
	fi; \
	cp "$$1" ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/terraform-provider-${NAME}

# Local install that builds for the current host and installs into both user plugins and the fixture
install-local:
	@echo "Building provider for local host and installing into user plugin dir and test fixture..."
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	# Build provider binary for host (uses default GOOS/GOARCH of the host)
	go build -o ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/terraform-provider-${NAME} ./
	chmod +x ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/terraform-provider-${NAME}
	# Also copy into the fixture-local plugin path used by tests (relative path kept)
	mkdir -p test/fixtures/full_e2e/plugins/arup.com/platform/sonarcloud/${VERSION}/${OS_ARCH}
	cp ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/terraform-provider-${NAME} test/fixtures/full_e2e/plugins/arup.com/platform/sonarcloud/${VERSION}/${OS_ARCH}/terraform-provider-sonarcloud
	chmod +x test/fixtures/full_e2e/plugins/arup.com/platform/sonarcloud/${VERSION}/${OS_ARCH}/terraform-provider-sonarcloud

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

# E2E tests using Terratest
# Requires: SONARCLOUD_ORGANIZATION, SONARCLOUD_TOKEN, SONARCLOUD_TEST_USER_LOGIN
# Note: Runs 'make install-local' first to ensure the provider is available locally
e2e: install-local
	go test -v -timeout 30m ./test/...

# E2E test with coverage
e2e-coverage:
	mkdir -p $(COVER_DIR)/e2e
	go test -v -timeout 30m $(COVER_FLAGS) -coverprofile=$(COVER_DIR)/e2e/coverage.out ./test/...

fmt:
	go fmt ./sonarcloud

docs:
	go generate ./...

docs-check: docs
	git diff --quiet

.PHONY: docs e2e e2e-coverage
