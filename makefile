export ARTIFACT:=feeder
export SHELL:=/usr/bin/env bash -O extglob -c
export GO111MODULE:=on
export OS:=$(shell uname | tr '[:upper:]' '[:lower:]')

clean:
	rm -fv ${ARTIFACT}

build: GOOS ?= ${OS}
build: GOARCH ?= amd64
build: clean
	GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -ldflags "-X main.buildTime=`date --iso-8601=s` -X main.buildVersion=`git rev-parse HEAD | cut -c-7`" .

release-linux:
	GOOS=linux $(MAKE) build
	tar Jcf ${ARTIFACT}-`git describe --abbrev=0 --tags`-linux-amd64.txz ${ARTIFACT}

release-darwin:
	GOOS=darwin $(MAKE) build
	tar Jcf ${ARTIFACT}-`git describe --abbrev=0 --tags`-darwin-amd64.txz ${ARTIFACT}

release: clean release-linux release-darwin

run: build
	./${ARTIFACT} -config ~/.config/feeder/config.yml

test: build
	go test -v -vet=all -failfast
