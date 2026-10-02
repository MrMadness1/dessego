GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BINARY ?= dessego
REVISION ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)

build:
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -mod=vendor -ldflags="-s -w -X main.buildRevision=$(REVISION)" -o bin/$(BINARY)-$(GOOS)-$(GOARCH) ./cmd/server

test:
	go test -race -mod=vendor ./...

vet:
	go vet -mod=vendor ./...

fmt-check:
	@test -z "$$(gofmt -l cmd internal)" || { gofmt -l cmd internal; exit 1; }

lint:
	golangci-lint run ./cmd/... ./internal/...

deps:
	go mod verify && go mod tidy && go mod vendor

docs:
	swagger generate spec -m -o swagger.yaml

.PHONY: build test vet fmt-check lint deps docs
