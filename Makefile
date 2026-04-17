SHELL := /bin/bash

DIST_DIR := dist
BUILD_DIR := $(DIST_DIR)/build
ASSETS_DIR := $(DIST_DIR)/assets

VERSION ?= dev
GO_BUILD_FLAGS := -trimpath

CLI_TARGET := ./cmd/commandline
SLIM_TARGET := ./cmd/slim

.PHONY: clean build-cli archive-cli package-cli build-slim archive-slim package-slim

clean:
	rm -rf $(DIST_DIR)

build-cli:
	@test -n "$(GOOS)" || (echo "GOOS is required" >&2; exit 1)
	@test -n "$(GOARCH)" || (echo "GOARCH is required" >&2; exit 1)
	@binary_name="dify-plugin"; \
		package_name="dify-plugin-$(GOOS)-$(GOARCH)"; \
		stage_dir="$(BUILD_DIR)/cli/$(GOOS)-$(GOARCH)/$$package_name"; \
		if [[ "$(GOOS)" == "windows" ]]; then \
			binary_name="$$binary_name.exe"; \
		fi; \
		rm -rf "$$stage_dir"; \
		mkdir -p "$$stage_dir"; \
		CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" \
			go build $(GO_BUILD_FLAGS) -ldflags "-X main.VersionX=$(VERSION)" \
			-o "$$stage_dir/$$binary_name" $(CLI_TARGET); \
		cp LICENSE "$$stage_dir/LICENSE"

archive-cli:
	@test -n "$(GOOS)" || (echo "GOOS is required" >&2; exit 1)
	@test -n "$(GOARCH)" || (echo "GOARCH is required" >&2; exit 1)
	@package_name="dify-plugin-$(GOOS)-$(GOARCH)"; \
		stage_root="$(BUILD_DIR)/cli/$(GOOS)-$(GOARCH)"; \
		stage_dir="$$stage_root/$$package_name"; \
		asset_root="$$(pwd)/$(ASSETS_DIR)"; \
		test -d "$$stage_dir" || (echo "missing build directory: $$stage_dir" >&2; exit 1); \
		mkdir -p "$$asset_root"; \
		if [[ "$(GOOS)" == "windows" ]]; then \
			asset_path="$$asset_root/$$package_name.zip"; \
			rm -f "$$asset_path"; \
			(cd "$$stage_root" && zip -rq "$$asset_path" "$$package_name"); \
		else \
			asset_path="$$asset_root/$$package_name.tar.gz"; \
			rm -f "$$asset_path"; \
			tar -C "$$stage_root" -czf "$$asset_path" "$$package_name"; \
		fi

package-cli: build-cli archive-cli

build-slim:
	@test -n "$(GOOS)" || (echo "GOOS is required" >&2; exit 1)
	@test -n "$(GOARCH)" || (echo "GOARCH is required" >&2; exit 1)
	@binary_name="dify-plugin-slim"; \
		package_name="dify-plugin-slim-$(GOOS)-$(GOARCH)"; \
		stage_dir="$(BUILD_DIR)/slim/$(GOOS)-$(GOARCH)/$$package_name"; \
		if [[ "$(GOOS)" == "windows" ]]; then \
			binary_name="$$binary_name.exe"; \
		fi; \
		rm -rf "$$stage_dir"; \
		mkdir -p "$$stage_dir"; \
		CGO_ENABLED=0 GOOS="$(GOOS)" GOARCH="$(GOARCH)" \
			go build $(GO_BUILD_FLAGS) -o "$$stage_dir/$$binary_name" $(SLIM_TARGET); \
		cp LICENSE "$$stage_dir/LICENSE"

archive-slim:
	@test -n "$(GOOS)" || (echo "GOOS is required" >&2; exit 1)
	@test -n "$(GOARCH)" || (echo "GOARCH is required" >&2; exit 1)
	@package_name="dify-plugin-slim-$(GOOS)-$(GOARCH)"; \
		stage_root="$(BUILD_DIR)/slim/$(GOOS)-$(GOARCH)"; \
		stage_dir="$$stage_root/$$package_name"; \
		asset_root="$$(pwd)/$(ASSETS_DIR)"; \
		test -d "$$stage_dir" || (echo "missing build directory: $$stage_dir" >&2; exit 1); \
		mkdir -p "$$asset_root"; \
		if [[ "$(GOOS)" == "windows" ]]; then \
			asset_path="$$asset_root/$$package_name.zip"; \
			rm -f "$$asset_path"; \
			(cd "$$stage_root" && zip -rq "$$asset_path" "$$package_name"); \
		else \
			asset_path="$$asset_root/$$package_name.tar.gz"; \
			rm -f "$$asset_path"; \
			tar -C "$$stage_root" -czf "$$asset_path" "$$package_name"; \
		fi

package-slim: build-slim archive-slim
