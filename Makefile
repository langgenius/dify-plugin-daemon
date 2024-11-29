# Makefile for building dify-plugin
OS := darwin
ARCHS := amd64 arm64
BIN_DIR := bin
CMD_DIR := ./cmd/commandline
TAR_EXT := tar.gz

.PHONY: build
build:
	@for arch in $(ARCHS); do \
		GOOS=$(OS) GOARCH=$$arch go build -o $(BIN_DIR)/dify-plugin-$(OS)-$$arch $(CMD_DIR); \
		chmod +x $(BIN_DIR)/dify-plugin-$(OS)-$$arch; \
	done

.PHONY: tarball
tarball: build
	@for arch in $(ARCHS); do \
		tar -czvf $(BIN_DIR)/dify-plugin-$(OS)-$$arch.$(TAR_EXT) -C $(BIN_DIR) dify-plugin-$(OS)-$$arch; \
	done

.PHONY: sha256
sha256: tarball
	@for arch in $(ARCHS); do \
		shasum -a 256 $(BIN_DIR)/dify-plugin-$(OS)-$$arch.$(TAR_EXT) | awk '{ print $$1 }' > $(BIN_DIR)/sha256_$(OS)_$$arch; \
	done

.PHONY: update_formula
update_formula: sha256
	@for arch in $(ARCHS); do \
		sed -i.bak \
			-e "s/sha256 \"SHA256_$(OS)_$${arch^^}\"/sha256 \"$(shell cat $(BIN_DIR)/sha256_$(OS)_$$arch)\"/g" \
			dify.rb; \
	done
	rm -f dify.rb.bak

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/* $(BIN_DIR)/sha256_*

.PHONY: all
all: update_formula