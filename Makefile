.PHONY: help build run test tidy vet swagger clean

APP_DIR := api
BIN_DIR := $(APP_DIR)/bin

## help: show available targets
help:
	@printf '\nAvailable targets:\n'
	@grep -E '^##' Makefile | sed 's/## //' | column -t -s ':' | sed 's/^/  /'
	@printf '\n'

## build: compile the url-shortner API binary into bin/
build:
	cd $(APP_DIR) && go build -o bin/server ./cmd/main

## run: run the API server on :8080
run:
	cd $(APP_DIR) && go run ./cmd/main

## test: run all tests
test:
	cd $(APP_DIR) && go test ./...

## vet: run go vet
vet:
	cd $(APP_DIR) && go vet ./...

## tidy: tidy go modules
tidy:
	cd $(APP_DIR) && go mod tidy

## clean: remove built binaries
clean:
	rm -rf $(BIN_DIR)

## install: build and move binary into bin/
install: build
	@echo "binary built: $(BIN_DIR)/server"