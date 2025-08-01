PACKAGE=stamus-ctl/internal/app
LOGGER=stamus-ctl/internal/logging

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=stamusctl
DAEMON_NAME=stamusd


HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

TARGET_ARCH?=linux/amd64

VERSION:=$(if $(VERSION),$(VERSION),$(shell git describe --tags --abbrev=0))
GIT_COMMIT:=$(if $(GIT_COMMIT),$(GIT_COMMIT),$(shell git rev-parse HEAD))

GOPATH?=$(shell if test -x `which go`; then go env GOPATH; else echo "$(HOME)/go"; fi)
GOCACHE?=$(HOME)/.cache/go-build


STATIC_BUILD?=true

DEV_IMAGE?=false

override LDFLAGS += \
  -X ${PACKAGE}.Arch=${TARGET_ARCH} \
  -X ${PACKAGE}.Commit=${GIT_COMMIT} \
  -X ${PACKAGE}.Version=${VERSION} \
  -X ${LOGGER}.envType=prd \
  -extldflags=-static

all: cli daemon

cli:
	CGO_ENABLED=0 go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${CLI_NAME} ./cmd

test-cli:
	CGO_ENABLED=0 BUILD_MODE=test STAMUS_APP_NAME=stamusctl go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${CLI_NAME} ./cmd

test:
	CGO_ENABLED=0 go test ./... -cover

cover:
	go test -coverprofile=profile.cov.tmp ./...
	cat profile.cov.tmp | grep -v "_mocks.go" | grep -v "stamus-ctl/cmd" | grep -v "main.go" | grep -v "docs.go" > cover.out
	go tool cover -func cover.out

daemon:
	CGO_ENABLED=0 go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${DAEMON_NAME} ./cmd

daemon-dev:
	air run

daemon-test:
	go test ./.test

build-swaggo-image:
	docker build . -t swag-daemon -f docker/Dockerfile.swag

update-swagger: build-swaggo-image
	docker run --rm -it -v .:/code swag-daemon:latest

lint:
	golangci-lint run --timeout 5m

fmt:
	GOFUMPT_SPLIT_LONG_LINES="on" gofumpt -w -s .
	goimports -w .

fmt-check:
	GOFUMPT_SPLIT_LONG_LINES="on" gofumpt -l .
	goimports -l .

# This step is needed in tests to have embeds loaded in some xdg paths
init-embeds:
	STAMUS_APP_NAME=stamusctl EMBED_MODE=true go run ./cmd compose init -h

.PHONY: all cli test-cli test daemon daemon-dev daemon-test build-swaggo-image update-swagger init-embeds
