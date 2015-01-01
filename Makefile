# GOOS=linux GOARCH=amd64

BINARIES := builder cache controller database registry router
BINARY_DEST_DIR := bin

all: build test

build:
	for i in $(BINARIES); do \
		mkdir -p $(BINARY_DEST_DIR)/$$i; \
	  CGO_ENABLED=0 godep go build -a -v -ldflags '-s' -o $(BINARY_DEST_DIR)/$$i/boot examples/$$i/boot.go; \
	done

test: test-unit

test-unit:
	godep go test -v ./...
