# MIT License
#
# (C) Copyright 2021-2022,2025 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Service
NAME       ?= smd
GIT_STATE  := $(shell if git diff-index --quiet HEAD --; then echo 'clean'; else echo 'dirty'; fi)
BUILD_HOST := $(shell hostname)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := $(shell go version | awk '{print $3}')
BUILD_USER := $(shell whoami)
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT     := $(shell git rev-parse HEAD)
VERSION    ?= $(shell git describe --tags --always --abbrev=0)
VERSION_D  := $(shell git describe --tags --always --abbrev=0 --dirty --broken)
LDFLAGS    := -ldflags "-X main.GitCommit=$(COMMIT) \
	-X 'main.BuildTime=$(BUILD_TIME)' \
	-X 'main.Version=$(VERSION)' \
	-X 'main.GitBranch=$(BRANCH)' \
	-X 'main.GitTag=$(VERSION)' \
	-X 'main.GitState=$(GIT_STATE)' \
	-X 'main.BuildHost=$(BUILD_HOST)' \
	-X 'main.GoVersion=$(GO_VERSION)' \
	-X 'main.BuildUser=$(BUILD_USER)'"

# RPM version/release: strip the leading 'v' and drop git-describe's
# '-N-gHASH[-dirty]' suffix (hyphens aren't allowed in an RPM Version
# field anyway). An exact tag like v0.1.2 becomes 0.1.2.
RPM_GIT_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
RPM_VERSION ?= $(shell echo "$(RPM_GIT_VERSION)" | sed -e 's/^v//' -e 's/-.*//')
RPM_RELEASE ?= 1
RPM_TOPDIR ?= $(CURDIR)/dist/rpmbuild
RPM_NAME ?= smd-quadlet


all: binaries binaries-pprof image image-pprof ct-image unittest

.PHONY : all image image-pprof unittest ct-image binaries binaries-pprof coverage docker rpm-build rpm-clean

image:
	docker build $(NO_CACHE) --pull $(DOCKER_ARGS) --tag '$(NAME):$(VERSION)' -f Dockerfile .

unittest:
	go test -cover -v -tags musl ./...

ct-image:
	docker build --no-cache -f test/Dockerfile test/ --tag smd-test:$(VERSION)

binaries: smd smd-init smd-loader native

smd: cmd/smd/*.go
	GOOS=linux GOARCH=amd64 go build -o smd -v -tags musl $(LDFLAGS) ./cmd/smd

smd-init: cmd/smd-init/*.go
	GOOS=linux GOARCH=amd64 go build -o smd-init -v -tags musl $(LDFLAGS) ./cmd/smd-init

smd-loader: cmd/smd-loader/*.go
	GOOS=linux GOARCH=amd64 go build -o smd-loader -v -tags musl $(LDFLAGS) ./cmd/smd-loader


native:
	go build -o smd-init-native -v -tags musl $(LDFLAGS) ./cmd/smd-init
	go build -o smd-native -v -tags musl $(LDFLAGS) ./cmd/smd
	go build -o smd-loader-native -v -tags musl $(LDFLAGS) ./cmd/smd-loader

coverage:
	go test -cover -v -tags musl ./cmd/* ./internal/* ./pkg/*

binaries-pprof: pprof/smd pprof/smd-init pprof/smd-loader

pprof/smd: cmd/smd/*.go
	GOOS=linux GOARCH=amd64 go build -o pprof/smd -v -tags "musl pprof" $(LDFLAGS) ./cmd/smd

pprof/smd-init: cmd/smd-init/*.go
	GOOS=linux GOARCH=amd64 go build -o pprof/smd-init -v -tags "musl pprof" $(LDFLAGS) ./cmd/smd-init

pprof/smd-loader: cmd/smd-loader/*.go
	GOOS=linux GOARCH=amd64 go build -o pprof/smd-loader -v -tags "musl pprof" $(LDFLAGS) ./cmd/smd-loader

image-pprof:
	docker build ${NO_CACHE} --pull ${DOCKER_ARGS} --tag '${NAME}-pprof:${VERSION}' -f Dockerfile.pprof .

clean:
	rm -f smd smd-init smd-init-native smd-loader smd-loader-native smd-native
	rm -rf pprof
	go clean -testcache
	go clean -cache
	go clean -modcache

docker: smd smd-init smd-loader
	docker build -t ghcr.io/openchami/smd:$(VERSION_D) .

rpm-build: ## Build the smd RPM (VERSION/RPM_RELEASE override the derived defaults)
	@command -v rpmbuild >/dev/null 2>&1 || { echo "rpmbuild is required but not installed."; exit 1; }
	rm -rf $(RPM_TOPDIR)
	mkdir -p $(RPM_TOPDIR)/SOURCES/$(RPM_NAME)-$(RPM_VERSION)/LICENSES
	cp packaging/rpm-quadlet/systemd/* $(RPM_TOPDIR)/SOURCES/$(RPM_NAME)-$(RPM_VERSION)/
	cp LICENSES/MIT.txt $(RPM_TOPDIR)/SOURCES/$(RPM_NAME)-$(RPM_VERSION)/LICENSES/
	tar -C $(RPM_TOPDIR)/SOURCES -czf $(RPM_TOPDIR)/SOURCES/$(RPM_NAME)-$(RPM_VERSION).tar.gz \
		$(RPM_NAME)-$(RPM_VERSION)
	rpmbuild --define "_topdir $(RPM_TOPDIR)" \
		--define "version $(RPM_VERSION)" \
		--define "rel $(RPM_RELEASE)" \
		-bb packaging/rpm-quadlet/smd-quadlet.spec
	@echo "Built: $(RPM_TOPDIR)/RPMS/noarch/$$(ls $(RPM_TOPDIR)/RPMS/noarch)"

rpm-clean: ## Remove local RPM build artifacts
	rm -rf $(RPM_TOPDIR)

