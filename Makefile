BINARY := stxai
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

build-all: clean
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-windows-amd64.exe .
	@echo "Built all targets in bin/"

install:
	go build -ldflags "$(LDFLAGS)" -o /usr/local/bin/$(BINARY) .

clean:
	rm -rf bin/

.PHONY: build build-all install clean
