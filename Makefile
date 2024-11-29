.PHONY: build
build:
	GOOS=darwin GOARCH=amd64 go build -o bin/dify-plugin-darwin-amd64 ./cmd/commandline
	GOOS=darwin GOARCH=arm64 go build -o bin/dify-plugin-darwin-arm64 ./cmd/commandline
	chmod +x bin/dify-plugin-darwin-amd64
	chmod +x bin/dify-plugin-darwin-arm64

.PHONY: tarball
tarball: build
	tar -czvf bin/dify-plugin-darwin-amd64.tar.gz -C bin dify-plugin-darwin-amd64
	tar -czvf bin/dify-plugin-darwin-arm64.tar.gz -C bin dify-plugin-darwin-arm64

.PHONY: sha256
sha256: tarball
	shasum -a 256 bin/dify-plugin-darwin-amd64.tar.gz | awk '{ print $$1 }' > bin/sha256_darwin_amd64
	shasum -a 256 bin/dify-plugin-darwin-arm64.tar.gz | awk '{ print $$1 }' > bin/sha256_darwin_arm64

.PHONY: update_formula
update_formula: sha256
	sed -i.bak \
        -e "s/sha256 \"SHA256_DARWIN_AMD64\"/sha256 \"$(shell cat bin/sha256_darwin_amd64)\"/g" \
        -e "s/sha256 \"SHA256_DARWIN_ARM64\"/sha256 \"$(shell cat bin/sha256_darwin_arm64)\"/g" \
        dify.rb
	rm -f dify.rb.bak

.PHONY: clean
clean:
	rm -rf bin/* bin/sha256_*

.PHONY: all
all: update_formula