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
			hash=$$(cat $(BIN_DIR)/sha256_$$platform\_$$arch); \
			if [ "$$platform" = "darwin" ]; then \
				if [ "$$arch" = "amd64" ]; then \
					url="file://$$(pwd)/$(BIN_DIR)/dify-plugin-darwin-amd64.tar.gz"; \
					placeholder="sha256 \"3b0172bfdaf19396a855974b6f83e03a86ce2a073615cd7d6fbbb104c3d96946\""; \
				elif [ "$$arch" = "arm64" ]; then \
					url="file://$$(pwd)/$(BIN_DIR)/dify-plugin-darwin-arm64.tar.gz"; \
					placeholder="sha256 \"8a527f7bc61046aa11992d76cc2e3fe2a2c38cf3434d882273fcba30dd3a2e00\""; \
				fi; \
				echo "Updating formula for $$platform $$arch"; \
				sed -i '' "s|url \"file://.*$$platform-$$arch.tar.gz\"|url \"$$url\"|" dify.rb; \
				sed -i '' "s|$$placeholder|sha256 \"$$hash\"|" dify.rb; \
			fi; \
		done; \
	done
	@rm -f dify.rb.bak

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/*

.PHONY: all
all: update_formula
