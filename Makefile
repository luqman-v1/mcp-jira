.PHONY: build test check install clean

BINARY_NAME=mcp-jira

build:
	go build -o bin/$(BINARY_NAME) .

test:
	go test -v -count=1 ./...

check:
	go vet ./...
	go test -race -count=1 ./...
install:
	mkdir -p $(HOME)/go/bin
	go build -o $(HOME)/go/bin/$(BINARY_NAME) .

clean:
	rm -rf bin/
