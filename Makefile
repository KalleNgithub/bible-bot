export PATH := $(HOME)/.local/go/bin:$(PATH)

.PHONY: build test data

build:
	go build ./...

test:
	go test ./...

data:
	go run ./cmd/build-data
