.PHONY: all clean binaries build install

BINARY_NAME := bitrise-cache
BIN_DIR := bin

# Platforms to build
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Map Go GOOS/GOARCH to our naming convention
define build_binary
	$(eval GOOS := $(word 1,$(subst /, ,$1)))
	$(eval GOARCH := $(word 2,$(subst /, ,$1)))
	$(eval OS_NAME := $(if $(filter linux,$(GOOS)),Linux,$(if $(filter darwin,$(GOOS)),Darwin,Windows)))
	$(eval ARCH_NAME := $(if $(filter amd64,$(GOARCH)),x86_64,arm64))
	$(eval EXT := $(if $(filter windows,$(GOOS)),.exe,))
	@echo "Building for $(OS_NAME)/$(ARCH_NAME)..."
	@GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -mod=vendor -ldflags="-s -w" \
		-o $(BIN_DIR)/$(BINARY_NAME)-$(OS_NAME)-$(ARCH_NAME)$(EXT) .
endef

all: build

# Build JS
build:
	npm run build

# Install npm dependencies
install:
	npm install

# Clean binaries
clean-bin:
	rm -rf $(BIN_DIR)

# Clean all build artifacts
clean: clean-bin
	rm -rf dist lib node_modules dist-goreleaser

# Build for current platform only (for local testing)
build-local:
	@mkdir -p $(BIN_DIR)
	go build -mod=vendor -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) .
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

goreleaser:
	go run github.com/goreleaser/goreleaser/v2@latest release

goreleaser-snapshot:
	go run github.com/goreleaser/goreleaser/v2@latest release --auto-snapshot --clean --skip publish
