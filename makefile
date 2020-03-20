export SHELL:=/usr/bin/env bash -O extglob -c
export GO111MODULE:=on
export OS=$(shell uname | tr '[:upper:]' '[:lower:]')

build: GOOS ?= ${OS}
build: GOARCH ?= amd64
build:
	rm -f mugo
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags "-X main.buildTime=`date --iso-8601=s` -X main.buildVersion=`git rev-parse HEAD | cut -c-7`" .

release-linux: testing
	GOOS=linux $(MAKE) build
	tar Jcf mugo-`git describe --abbrev=0 --tags`-linux-amd64.txz mugo

release-darwin:
	GOOS=darwin $(MAKE) build
	tar Jcf mugo-`git describe --abbrev=0 --tags`-darwin-amd64.txz mugo

release: test clean release-linux release-darwin

test: clean
	go test -v -vet=all -failfast

clean:
	rm -f mugo
	rm -f mugo-*.txz

run: build
	./mugo
