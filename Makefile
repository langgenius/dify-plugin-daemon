PLATFORMS = darwin linux windows
ARCHS = amd64 arm64
BIN_DIR = bin
CMD_DIR = ./cmd/commandline

.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			if [ "$$platform" = "windows" ]; then \
				ext=".exe"; \
			else \
				ext=""; \
			fi; \
			bin_name=dify-plugin-$$platform-$$arch$$ext; \
			echo "Building $$bin_name"; \
			GOOS=$$platform GOARCH=$$arch go build -o $(BIN_DIR)/$$bin_name $(CMD_DIR); \
			if [ "$$platform" != "windows" ]; then chmod +x $(BIN_DIR)/$$bin_name; fi; \
		done; \
	done

.PHONY: tarball
tarball: build
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			if [ "$$platform" = "windows" ]; then \
				ext=".exe"; \
				archive=$(BIN_DIR)/dify-plugin-$$platform-$$arch.zip; \
				echo "Creating $$archive"; \
				zip -j $$archive $(BIN_DIR)/dify-plugin-$$platform-$$arch$$ext; \
			else \
				ext=""; \
				archive=$(BIN_DIR)/dify-plugin-$$platform-$$arch.tar.gz; \
				echo "Creating $$archive"; \
				tar -czvf $$archive -C $(BIN_DIR) dify-plugin-$$platform-$$arch$$ext; \
			fi; \
		done; \
	done

.PHONY: sha256
sha256: tarball
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			if [ "$$platform" = "windows" ]; then \
				archive=$(BIN_DIR)/dify-plugin-$$platform-$$arch.zip; \
			else \
				archive=$(BIN_DIR)/dify-plugin-$$platform-$$arch.tar.gz; \
			fi; \
			hash_file=$(BIN_DIR)/sha256_$$platform\_$$arch; \
			echo "Computing SHA256 for $$archive"; \
			shasum -a 256 $$archive | awk '{ print $$1 }' > $$hash_file; \
			echo "SHA256: $$(cat $$hash_file)"; \
		done; \
	done
.PHONY: update-brewfile
update-brewfile: sha256
	@echo "Updating dify.rb"
	@amd64_checksum=$$(cat $(BIN_DIR)/sha256_darwin_amd64); \
	arm64_checksum=$$(cat $(BIN_DIR)/sha256_darwin_arm64); \
	sed -e "s/PLACEHOLDER_FOR_AMD64_CHECKSUM/$$amd64_checksum/" \
		-e "s/PLACEHOLDER_FOR_ARM64_CHECKSUM/$$arm64_checksum/" \
		dify.rb.template > dify.rb

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/*
	
.PHONY: all
all: clean update-brewfile