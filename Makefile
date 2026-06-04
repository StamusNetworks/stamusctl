PACKAGE=stamus-ctl/internal/app
LOGGER=stamus-ctl/internal/logging

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=stamusctl
DAEMON_NAME=stamusd


HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

TARGET_ARCH?=linux/amd64

VERSION:=$(if $(VERSION),$(VERSION),$(shell git describe --tags --abbrev=0 2>/dev/null || cat VERSION 2>/dev/null || echo "unknown"))
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

nix-test-syntax:
	@echo "Validating NixOS template syntax..."
	@find internal/embeds/ -name "*.nix" -exec nix-instantiate --parse {} \; 2>&1 || true

nix-test-vm:
	@echo "Running NixOS VM integration tests..."
	nix build .#checks.$(shell nix eval --raw 'builtins.currentSystem').nixos-test

nix-test-vm-docker:
	@echo "Running NixOS VM integration tests in Docker..."
	docker run --rm \
		--device /dev/kvm \
		-v $(CURRENT_DIR):/src \
		-w /src \
		nixos/nix:latest \
		sh -c "echo 'extra-experimental-features = nix-command flakes' >> /etc/nix/nix.conf \
			&& git config --global --add safe.directory /src \
			&& nix build .#checks.x86_64-linux.nixos-test -L"

nix-test-cmd-docker:
	@echo "Running stamusctl nix init + test in NixOS Docker container..."
	docker run --rm \
		-v $(CURRENT_DIR):/src:ro \
		nixos/nix:latest \
		sh -c '\
			echo "extra-experimental-features = nix-command flakes" >> /etc/nix/nix.conf \
			&& echo "Building stamusctl..." \
			&& cd /tmp && cp -r /src stamusctl && cd stamusctl \
			&& git config --global --add safe.directory /tmp/stamusctl \
			&& nix build -L \
			&& echo "Running nix init with test fixture template..." \
			&& mkdir -p /tmp/stamus-templates/clearndr/embedded/ \
			&& cp -r tests/nixos/fixtures/template/* /tmp/stamus-templates/clearndr/embedded/ \
			&& EMBED_MODE=true STAMUS_TEMPLATES_FOLDER=/tmp/stamus-templates/ \
			   STAMUS_APP_NAME=stamusctl \
			   ./result/bin/stamusctl nix init --config /tmp/test-config --default \
			&& echo "Running nix test on rendered config..." \
			&& touch /etc/NIXOS \
			&& STAMUS_APP_NAME=stamusctl \
			   ./result/bin/stamusctl nix test --config /tmp/test-config \
		'

nix-iso:
	@echo "Building NixOS live ISO with stamusctl..."
	nix build .#packages.x86_64-linux.iso -L
	@echo "ISO built: $$(readlink -f result)/iso/stamusctl-live.iso"

nix-iso-run: nix-iso
	@echo "Launching ISO in QEMU..."
	qemu-system-x86_64 \
		-enable-kvm \
		-m 4G \
		-smp 2 \
		-cdrom $$(find $$(readlink -f result)/iso/ -name '*.iso' | head -1) \
		-boot d \
		-vga virtio \
		-display gtk \
		-usb -device usb-tablet \
		-nic user,model=virtio-net-pci

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

# Shell completion scripts
completions:
	@mkdir -p completions
	@echo "Generating bash completion..."
	@STAMUS_APP_NAME=stamusctl go run ./cmd completion bash > completions/stamusctl.bash
	@echo "Generating zsh completion..."
	@STAMUS_APP_NAME=stamusctl go run ./cmd completion zsh > completions/stamusctl.zsh
	@echo "Generating fish completion..."
	@STAMUS_APP_NAME=stamusctl go run ./cmd completion fish > completions/stamusctl.fish
	@echo "Generating powershell completion..."
	@STAMUS_APP_NAME=stamusctl go run ./cmd completion powershell > completions/stamusctl.ps1
	@echo "Completions generated in completions/"

# Man pages
man-pages:
	@mkdir -p docs/man
	@echo "Generating man pages..."
	@STAMUS_APP_NAME=stamusctl go run ./cmd/doc
	@echo "Man pages generated in docs/man/"

# Install completion scripts to system directories
install-completions: completions
	@echo "Installing bash completion..."
	@if [ -d /etc/bash_completion.d ]; then \
		sudo cp completions/stamusctl.bash /etc/bash_completion.d/stamusctl; \
		echo "Installed to /etc/bash_completion.d/stamusctl"; \
	elif [ -d /usr/local/etc/bash_completion.d ]; then \
		sudo cp completions/stamusctl.bash /usr/local/etc/bash_completion.d/stamusctl; \
		echo "Installed to /usr/local/etc/bash_completion.d/stamusctl"; \
	else \
		echo "No bash completion directory found. Copy completions/stamusctl.bash manually."; \
	fi
	@echo "For zsh, copy completions/stamusctl.zsh to a directory in your fpath"
	@echo "For fish, copy completions/stamusctl.fish to ~/.config/fish/completions/"

.PHONY: all cli test-cli test daemon daemon-dev daemon-test nix-test-syntax nix-test-vm nix-test-vm-docker nix-test-cmd-docker nix-iso nix-iso-run build-swaggo-image update-swagger init-embeds completions man-pages install-completions
