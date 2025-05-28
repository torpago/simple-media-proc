# Makefile for simple-media-proc

GIT_COMMIT ?= $(shell git rev-parse HEAD)
GIT_COMMIT_SHORT ?= $(shell git rev-parse --short HEAD)
LDFLAGS = "-X main.Version=$(GIT_COMMIT)"

# ImageMagick configuration
CGO_CFLAGS_ALLOW = -Xpreprocessor
PKG_CONFIG_PATH ?= /opt/homebrew/Cellar/imagemagick/7.1.0-51
LIBRARY_PATH ?= /opt/homebrew/Cellar/imagemagick/7.1.0-51/lib
CGO_ENABLED = 1

export CGO_CFLAGS_ALLOW PKG_CONFIG_PATH LIBRARY_PATH CGO_ENABLED

.PHONY: all test clean

all: test

test:
	go test -v ./...

clean:
	go clean
	rm -rf dist/

# Create test data directory
test-setup:
	mkdir -p test/data

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
