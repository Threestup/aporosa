BIN=aporosa
RELEASE?=$(shell git describe --abbrev=0 --tags)
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
PROJECT?=github.com/Threestup/aporosa
GOOS_?=linux
GOARCH_?=amd64

build: clean ## build the service
	go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${BIN}

release: clean ## build the service
	GOOS=${GOOS_} GOARCH=${GOARCH_} go build -tags netgo -a -v \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${BIN}

clean: ## clean project (generated binary + docker images)
	rm -rf ${BIN}

run: build
	./${BIN}

test:
	go test -v -race ./...

.PHONY: help migrations

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.SILENT:
