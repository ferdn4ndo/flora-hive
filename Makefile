.PHONY: build run test test-race test-docker lint fmt migrate-up migrate-down docker-build image

BIN := bin/flora-hive
BUILD_IMAGE ?= flora-hive:build
# Pin here and in Dockerfile ARG GO_VERSION (single place to document; pass --build-arg to override).
GO_VERSION ?= 1.24.4
DOCKER_BUILD_ARGS := --build-arg GO_VERSION=$(GO_VERSION)

# Produce bin/flora-hive using only the Go toolchain inside the image (see Dockerfile stages).
build:
	docker build $(DOCKER_BUILD_ARGS) --target build -t $(BUILD_IMAGE) .
	mkdir -p bin
	cid=$$(docker create $(BUILD_IMAGE)); \
	  docker cp $$cid:/out/flora-hive $(BIN) && docker rm -v $$cid
	chmod +x $(BIN)

run: build
	./$(BIN) app:serve

# Host Go (fast iteration when you have a local toolchain matching go.mod).
test:
	go test ./...

test-race:
	go test -race -count=1 ./...

# Tests using the same pinned Go image as release builds (no host Go required).
test-docker:
	docker build $(DOCKER_BUILD_ARGS) --target test .

lint:
	golangci-lint run

fmt:
	golangci-lint run --fix

migrate-up: build
	./$(BIN) migrate:up

migrate-down: build
	./$(BIN) migrate:down

# Runtime image (includes migrations + entrypoint).
docker-build: image

image:
	docker build $(DOCKER_BUILD_ARGS) -t flora-hive:local .
