PACKAGE=stamus-ctl/internal/app
LOGGER=stamus-ctl/internal/logging

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=stamusctl
DAEMON_NAME=stamusd


HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

TARGET_ARCH?=linux/amd64

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)
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

daemon:
	go build -v -ldflags '${LDFLAGS}' -o ${DIST_DIR}/${DAEMON_NAME} ./cmd

daemon-dev:
	air run

daemon-test: init-embeds
	EMBED_MODE=true go test ./.test/unit

build-swaggo-image:
	docker build . -t swag-daemon -f docker/Dockerfile.swag

update-swagger: build-swaggo-image
	docker run --rm -it -v .:/code swag-daemon:latest

# This step is needed in tests to have embeds loaded in some xdg paths
init-embeds:
	STAMUS_APP_NAME=stamusctl EMBED_MODE=true go run ./cmd compose init -h

.PHONY: all cli test-cli test daemon daemon-dev daemon-test build-swaggo-image update-swagger init-embeds
