.DEFAULT_GOAL := lint

CONTAINER_ENGINE ?= podman
OUTPUT_DIR = out

clean:
	rm -rf *.test

GOFLAGS=-mod=mod
build:
	CGO_ENABLED=0 go test -v -c

build-image:
	${CONTAINER_ENGINE} build -t validate-mcsc-upgrades:latest .

fmt:
	gofmt -s -w .

lint: fmt
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	golangci-lint run .
