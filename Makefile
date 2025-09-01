# Copyright (c) 2024 Alibaba Group Holding Ltd.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#-------------------------------------------------------------------------------
# General build options
MAIN_VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')

CURRENT_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
CURRENT_ARCH := $(shell uname -m | sed 's/aarch64/arm64/;s/armv7l/arm/;s/armv6l/arm/')

MOD_NAME := github.com/alibaba/loongsuite-go-agent
STRIP_DEBUG := -s -w

OUTPUT_BASE = otel
OUTPUT_DARWIN_AMD64 = $(OUTPUT_BASE)-darwin-amd64
OUTPUT_LINUX_AMD64 = $(OUTPUT_BASE)-linux-amd64
OUTPUT_WINDOWS_AMD64 = $(OUTPUT_BASE)-windows-amd64.exe
OUTPUT_DARWIN_ARM64 = $(OUTPUT_BASE)-darwin-arm64
OUTPUT_LINUX_ARM64 = $(OUTPUT_BASE)-linux-arm64

#-------------------------------------------------------------------------------
# Prepare version
# Get the current Git commit ID
CHECK_GIT_DIRECTORY := $(if $(wildcard .git),true,false)
ifeq ($(CHECK_GIT_DIRECTORY),true)
	COMMIT_ID := $(shell git rev-parse --short HEAD)
else
	COMMIT_ID := default
endif

VERSION := $(MAIN_VERSION)_$(COMMIT_ID)
XVALUES := -X=$(MOD_NAME)/tool/config.ToolVersion=$(VERSION) \
		   -X=$(MOD_NAME)/pkg/inst-api/version.Tag=v$(VERSION)

LDFLAGS := -ldflags="$(XVALUES) $(STRIP_DEBUG)"
BUILD_CMD = CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -a -trimpath $(LDFLAGS) -o $(3) ./tool/otel
BUILD_CMD_DEV = CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -a $(LDFLAGS) -o $(3) ./tool/otel

#-------------------------------------------------------------------------------
# Multiple OS and ARCH support
ifeq ($(CURRENT_ARCH), x86_64)
   CURRENT_ARCH := amd64
endif

# Check if current os contains "MINGW" or "MSYS" to determine if it is Windows
ifeq ($(findstring mingw,$(CURRENT_OS)),mingw)
   CURRENT_OS := windows
endif

ifeq ($(findstring msys,$(CURRENT_OS)),msys)
   CURRENT_OS := windows
endif

#-------------------------------------------------------------------------------
# Build targets
.PHONY: build
build: package-pkg tidy lint
	@echo "Building $(OUTPUT_BIN)..."
	$(eval OUTPUT_BIN=$(OUTPUT_BASE))
ifeq ($(CURRENT_OS),windows)
	$(eval OUTPUT_BIN=$(OUTPUT_BASE).exe)
endif
	@$(call BUILD_CMD_DEV,$(CURRENT_OS),$(CURRENT_ARCH),$(OUTPUT_BIN))
	@echo "Built completed: $(OUTPUT_BIN) $(VERSION)"

.PHONY: all test clean

all: clean darwin_amd64 linux_amd64 windows_amd64 darwin_arm64 linux_arm64
	@echo "All builds completed: $(OUTPUT_DARWIN_AMD64) $(OUTPUT_LINUX_AMD64) $(OUTPUT_WINDOWS_AMD64) $(OUTPUT_DARWIN_ARM64) $(OUTPUT_LINUX_ARM64)"

darwin_amd64: package-pkg tidy
	@echo "Building darwin_amd64..."
	@$(call BUILD_CMD,darwin,amd64,$(OUTPUT_DARWIN_AMD64))

linux_amd64: package-pkg tidy
	@echo "Building linux_amd64..."
	@$(call BUILD_CMD,linux,amd64,$(OUTPUT_LINUX_AMD64))

windows_amd64: package-pkg tidy
	@echo "Building windows_amd64..."
	@$(call BUILD_CMD,windows,amd64,$(OUTPUT_WINDOWS_AMD64))

darwin_arm64: package-pkg tidy
	@echo "Building darwin_arm64..."
	@$(call BUILD_CMD,darwin,arm64,$(OUTPUT_DARWIN_ARM64))

linux_arm64: package-pkg tidy
	@echo "Building linux_arm64..."
	@$(call BUILD_CMD,linux,arm64,$(OUTPUT_LINUX_ARM64))

.PHONY: tidy
tidy:
	@echo "Tidying up dependencies..."
	@go mod tidy

clean:
	@echo "Cleaning up..."
	@rm -f $(OUTPUT_DARWIN_AMD64) $(OUTPUT_LINUX_AMD64) $(OUTPUT_WINDOWS_AMD64) $(OUTPUT_DARWIN_ARM64) $(OUTPUT_LINUX_ARM64) $(OUTPUT_BASE)
	@go clean

test:
	go test -a -timeout 50m -v $(MOD_NAME)/test

install: build
	@echo "Running install process..."
	@cp $(OUTPUT_BASE) /usr/local/bin/
	@echo "Installed at /usr/local/bin/$(OUTPUT_BASE)"

#-------------------------------------------------------------------------------
# Package pkg module
# Embed the pkg module into the otel binary during the build process. When the
# otel tool needs to use it, it can directly extract and utilize the embedded
# package, instead of downloading it from the internet (via go mod tidy or
# go mod download).
# Since we want exclude test dependencies, here we create a temporary directory
# to hold the pkg module, remove test files, and refresh the go.mod file.
# Finally, we package it into a gzipped tarball for embedding.
# The tarball will be placed in tool/data/ directory.
PKG_GZIP = alibaba-pkg.gz
PKG_TMP = pkg_tmp
.PHONY: package-pkg
package-pkg:
	@echo "Packaging pkg module..."
	@rm -rf $(PKG_TMP)
	@cp -a pkg $(PKG_TMP)
	@find $(PKG_TMP)/* -iname '*_test.go' -exec rm -rf {} \;
	@cd $(PKG_TMP) && go mod tidy
	@tar -czf $(PKG_GZIP) --exclude='*.log' --exclude='*.string' --exclude='*.pprof' --exclude='*.gz' $(PKG_TMP)
	@mv alibaba-pkg.gz tool/data/
	@rm -rf $(PKG_TMP)

#-------------------------------------------------------------------------------
# Linting with golangci-lint
.PHONY: lint
lint:
	@LINTER=""; \
	if [ -n "$$GOBIN" ]; then \
		LINTER="$$GOBIN/golangci-lint"; \
	elif [ -n "$$GOPATH" ]; then \
		LINTER="$$GOPATH/bin/golangci-lint"; \
	else \
		LINTER="$$HOME/go/bin/golangci-lint"; \
	fi; \
	if [ ! -x "$$LINTER" ]; then \
  		echo "golangci-lint not found, installing to $$LINTER..."; \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6; \
	fi; \
	echo "Running golangci-lint..."; \
	$$LINTER run --config .golangci.yml ./tool/...
