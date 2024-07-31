# MIT License
#
# (C) Copyright 2021-2022 Hewlett Packard Enterprise Development LP
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
NAME ?= cray-smd
VERSION ?= $(shell cat .version)

all: image unittest ct snyk ct_image

.PHONY : all image unittest snyk ct ct_image binaries coverage docker

image:
	docker build ${NO_CACHE} --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' -f Dockerfile .

unittest:
	./runUnitTest.sh

snyk:
	./runSnyk.sh

ct:
	./runCT.sh

ct_image:
	docker build --no-cache -f test/ct/Dockerfile test/ct/ --tag smd-test:${VERSION} 

binaries: smd smd-init smd-loader native



BUILD := `git rev-parse --short HEAD`
VERSION := `git describe --tags --abbrev=0`

GOOS := $(if $(GOOS),$(GOOS),linux)
GOARCH := $(if $(GOARCH),$(GOARCH),amd64)

LDFLAGS=-ldflags "-X=$(GIT)main.commit=$(BUILD) -X=$(GIT)main.version=$(VERSION) -X=$(GIT)main.date=$(shell date +%Y-%m-%d:%H:%M:%S)"

smd: cmd/smd/*.go
	go build -o smd -v -tags musl $(LDFLAGS) ./cmd/smd

smd-init: cmd/smd-init/*.go
	go build -o smd-init -v -tags musl $(LDFLAGS) ./cmd/smd-init

native:
	go build -o smd-init-native -v -tags musl $(LDFLAGS) ./cmd/smd-init
	go build -o smd-native -v -tags musl $(LDFLAGS) ./cmd/smd
	go build -o smd-loader-native -v -tags musl $(LDFLAGS) ./cmd/smd-loader



smd-loader: cmd/smd-loader/*.go
	go build -o smd-loader -v -tags musl $(LDFLAGS) ./cmd/smd-loader

coverage:
	go test -cover -v -tags musl ./cmd/* ./internal/* ./pkg/*

clean:
	rm -f smd smd-init smd-loader smd-native smd-init-native
	go clean -testcache
	go clean -cache
	go clean -modcache

docker: smd smd-init smd-loader
	docker build -t bikeshack/smd:$(VERSION)-dirty .
