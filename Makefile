.PHONY: build test clean install fmt vet

VERSION ?= dev
LDFLAGS := -ldflags "-X github.com/ciaranRoche/morpheus/cmd.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/morpheus .

test:
	go test ./... -v

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) .

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet
