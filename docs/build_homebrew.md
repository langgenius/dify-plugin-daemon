# Build homebrew

## Intallation

### Install homebrew

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

brew -v
```

### Build CLI tool

```bash
# Build the CLI tool from source
make build 
# Generate the sha256 checksum
make sha256 
# Update the formula with the new sha256 checksum
make update_formula
# Clean the build
make clean
# Execute all the above commands
make all 
```

### Install CLI tool

```bash
make all
brew install --build-from-source ./dify.rb
```