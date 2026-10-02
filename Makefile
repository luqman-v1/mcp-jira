.PHONY: build test check install clean cross

BINARY_NAME=mcp-jira
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "0.0.1")
LDFLAGS := -s -w -X 'main.version=$(VERSION)'

build:
	go build -ldflags="$(LDFLAGS)" -o bin/$(BINARY_NAME) .

test:
	go test -v -count=1 ./...

check:
	go vet ./...
	go test -race -count=1 ./...

install:
	mkdir -p $(HOME)/go/bin
	go build -ldflags="$(LDFLAGS)" -o $(HOME)/go/bin/$(BINARY_NAME) .

cross:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-arm64.exe .

clean:
	rm -rf bin/ dist/
