# Makefile to log workflow.

# Set defaults (see http://clarkgrubb.com/makefile-style-guide#prologue)
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:
.ONESHELL:

GOFLAGS ?=

.PHONY: install
install: test
	go install $(GOFLAGS) ./...

.PHONY: build
build: build-linux64 build-win32

.PHONY: build-linux64
build-linux64: bindata
	mkdir -p ./bin
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o ./bin/hindsite ./...

.PHONY: build-win32
build-win32: bindata
	mkdir -p ./bin
	GOOS=windows GOARCH=386 go build $(GOFLAGS) -o ./bin/hindsite.exe ./...

.PHONY: test
test: bindata
	go test $(GOFLAGS) ./...

.PHONY: clean
clean:
	go clean $(GOFLAGS) -i ./...

.PHONY: doc
doc: install
	cp -p README.md doc/content/index.md
	hindsite build doc -v

.PHONY: serve
serve: doc
	hindsite serve doc

.PHONY: blog
blog: install
	hindsite build ./examples/blog -v
	hindsite serve ./examples/blog -v

.PHONY: push
push:
	git push -u --tags origin master

./hindsite/bindata.go: ./examples/builtin/*
	cd ./hindsite && go-bindata -prefix ../examples/builtin/template/ ../examples/builtin/template/...

.PHONY: bindata
bindata: ./hindsite/bindata.go