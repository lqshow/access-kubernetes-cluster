SHELL := /bin/bash

ifeq ($(VERSION),)
VERSION := $(shell git describe --always --dirty)
endif

all: build

.PHONY: build
build:
	./script/build.sh

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	@echo Formatting
	@go fmt ./...

.PHONY: clean
clean: clean-bin

.PHONY: clean-bin
clean-bin:
	rm -rf bin .go