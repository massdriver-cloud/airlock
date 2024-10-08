INSTALL_PATH ?= /usr/local/bin

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MKFILE_DIR := $(dir $(MKFILE_PATH))

all.macos: clean generate install.macos
all.linux: clean generate install.linux

.PHONY: check
check: clean generate test ## Run tests and linter locally
	golangci-lint run

.PHONY: clean
clean:
	rm -rf bin/*

.PHONY: generate
generate: ${API_DIR}/zz_generated.go

.PHONY: test
test:
	go test ./... -cover

bin:
	mkdir bin

.PHONY: lint
lint:
	golangci-lint run

.PHONY: build.macos
build.macos: bin
	GOOS=darwin GOARCH=arm64 go build -o bin/airlock-darwin-arm64

.PHONY: build.linux
build.linux: bin
	GOOS=linux GOARCH=amd64 go build -o bin/airlock-linux-amd64

.PHONY: install.macos
install.macos: build.macos
	rm -f ${INSTALL_PATH}/airlock
	cp bin/airlock-darwin-arm64 ${INSTALL_PATH}/airlock

.PHONY: install.linux
install.linux: build.linux
	cp -f bin/airlock-linux-amd64 ${INSTALL_PATH}/airlock
