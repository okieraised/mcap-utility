APP_NAME := mcap-utility
OUTPUT_DIR := build
PLATFORMS := linux windows darwin
ARCHS := amd64 arm64
VERSION := 0.1.0

all: clean build-all

build-all:
	@mkdir -p $(OUTPUT_DIR)
	@echo "Building $(APP_NAME)"
	@for GOOS in $(PLATFORMS); do \
		for GOARCH in $(ARCHS); do \
			EXT=""; \
			if [ "$$GOOS" = "windows" ]; then EXT=".exe"; fi; \
			OUT_FILE="$(OUTPUT_DIR)/$(APP_NAME)-$(VERSION)-$$GOOS-$$GOARCH$$EXT"; \
			echo " -> $$GOOS/$$GOARCH: $$OUT_FILE"; \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="-s -w" -o $$OUT_FILE . || exit 1; \
		done; \
	done

build-local:
	@go build -ldflags="-s -w" -o $(APP_NAME) .

build-dev:
	@go build -ldflags="-s -w" -o $(APP_NAME) .

clean:
	@rm -rf $(OUTPUT_DIR)
	@rm -f $(APP_NAME)

.PHONY: all build-local build-dev build-all clean
