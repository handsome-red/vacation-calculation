# Makefile
.PHONY: test test-cover build run

test:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

build:
	go build -o bin/app ./app/exe

run:
	go run ./cmd/api

lint:
	golangci-lint run