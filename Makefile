BINARY_NAME ?= nostar
BUILD_DIR ?= bin
PKG ?= ./
GOOS ?= linux
GOARCH ?= amd64
LDFLAGS ?= -s -w -extldflags '-static'

.PHONY: build clean tidy fmt test staticcheck check setup

build:
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(PKG)

clean:
	rm -rf $(BUILD_DIR)
	go clean -modcache

fmt:
	@fmt_files=$$(gofmt -l .); \
	if [ -n "$$fmt_files" ]; then \
		echo "gofmt needed on:"; \
		echo "$$fmt_files"; \
		exit 1; \
	fi

test:
	go test -v ./...

staticcheck:
	staticcheck ./...

check: fmt test staticcheck

setup:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/spf13/cobra-cli@latest
