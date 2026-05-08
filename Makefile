TARGETS := $(shell ls scripts)
BUILD_IMAGE := gmsa-builder
CURDIR := $(shell pwd)
host_uid := $(shell id -u 2>/dev/null || echo 1000)
host_gid := $(shell id -g 2>/dev/null || echo 1000)

# Detect OS: on Windows, OS=Windows_NT is set by the environment
ifeq ($(OS),Windows_NT)
  DOCKERFILE := Dockerfile-windows.build
  DOCKER_SOCKET_MOUNT := -v //./pipe/docker_engine://./pipe/docker_engine
else
  DOCKERFILE := Dockerfile.build
  DOCKER_SOCKET_MOUNT := -v /var/run/docker.sock:/var/run/docker.sock
endif

RAW_HOST_ARCH := $(shell uname -m 2>/dev/null || echo amd64)
HOST_ARCH := $(if $(filter x86_64,$(RAW_HOST_ARCH)),amd64,$(if $(filter aarch64,$(RAW_HOST_ARCH)),arm64,$(RAW_HOST_ARCH)))

build-image:
	docker build -f $(DOCKERFILE) -t $(BUILD_IMAGE) --build-arg ARCH=$(HOST_ARCH) .

$(TARGETS): build-image
	docker run --rm \
		$(DOCKER_SOCKET_MOUNT) \
		-v $(CURDIR):/source \
		-e REPO -e TAG -e DRONE_TAG -e CROSS \
		-e "HOST_UID=${host_uid}" \
		-e "HOST_GID=${host_gid}" \
		$(BUILD_IMAGE) $@

.DEFAULT_GOAL := default

# Charts Build Scripts

pull-scripts:
	./scripts/charts-build-scripts/pull-scripts

rebase:
	./scripts/charts-build-scripts/rebase

remove:
	./scripts/charts-build-scripts/remove-asset

CHARTS_BUILD_SCRIPTS_TARGETS := prepare patch clean clean-cache charts list index unzip zip standardize template

$(CHARTS_BUILD_SCRIPTS_TARGETS):
	@./scripts/charts-build-scripts/pull-scripts
	@./bin/charts-build-scripts $@

.PHONY: $(TARGETS) $(CHARTS_BUILD_SCRIPTS_TARGETS)
