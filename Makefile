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

all: image image-pprof unittest ct snyk ct_image

.PHONY : all image unittest snyk ct ct_image binaries coverage docker

image:
	docker build $(NO_CACHE) --pull $(DOCKER_ARGS) --tag '$(NAME):$(VERSION)' -f Dockerfile .

unittest:
	go test -cover -v -tags musl ./...

snyk:
	./runSnyk.sh

ct:
	./runCT.sh

ct_image:
	docker build --no-cache -f test/ct/Dockerfile test/ct/ --tag smd-test:$(VERSION})

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
