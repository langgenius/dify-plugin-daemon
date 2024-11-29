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
		done; \
	done

.PHONY: update_formula
update_formula: sha256
	@cp dify.rb dify.rb.bak
	@for platform in $(PLATFORMS); do \
			for arch in $(ARCHS); do \
					placeholder="SHA256_$$(echo $$platform | tr a-z A-Z)_$$(echo $$arch | tr a-z A-Z)"; \
					hash=$$(cat $(BIN_DIR)/sha256_$$platform\_$$arch); \
					echo "Updating formula for $$placeholder"; \
					sed -i '' "s/sha256 \"$$placeholder\"/sha256 \"$$hash\"/" dify.rb; \
			done; \
	done
	@rm -f dify.rb.bak

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/*

.PHONY: all
all: update_formula
