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
build: bindata
	go build $(GOFLAGS) -o /tmp/hindsite ./...

.PHONY: test
test: bindata
	go test $(GOFLAGS) ./...

.PHONY: clean
clean:
	go clean $(GOFLAGS) -i ./...

.PHONY: doc
doc: install
	cp -p README.md doc/index.md
	cd doc
	hindsite build -content . -template . -v

.PHONY: serve
serve: doc
	hindsite serve -project doc

.PHONY: blog
blog: install
	hindsite build -project ./examples/blog -set urlprefix=/blog -v
	hindsite serve -project ./examples/blog -set urlprefix=/blog -v

.PHONY: push
push:
	git push -u --tags origin master

./hindsite/bindata.go: ./examples/builtin/*
	cd ./hindsite && go-bindata -prefix ../examples/builtin/template/ ../examples/builtin/template/...

.PHONY: bindata
bindata: ./hindsite/bindata.go