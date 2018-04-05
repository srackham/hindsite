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

BINDATA_FILES = $(shell find ./builtin/minimal/template) $(shell find ./builtin/blog/template)

./hindsite/bindata.go: $(BINDATA_FILES)
	cd ./hindsite
	go-bindata -prefix ../builtin/ -ignore '/(build|content)/' ../builtin/...

.PHONY: bindata
bindata: ./hindsite/bindata.go

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
doc: build-doc serve-doc

.PHONY: build-doc
build-doc: install
	cp -p README.md doc/content/index.md
	hindsite build doc

.PHONY: serve-doc
serve-doc: build-doc
	hindsite serve doc

.PHONY: push
push:
	git push -u --tags origin master

#
# Builtin blog development tasks.
#
BLOG_DIR = ./builtin/blog

.PHONY: blog
blog: build-blog serve-blog

# Built the builtin blog's init directory.
.PHONY: build-blog
build-blog: install
	hindsite build $(BLOG_DIR) -content $(BLOG_DIR)/template/init

.PHONY: serve-blog
serve-blog: build-blog
	hindsite serve $(BLOG_DIR)

.PHONY: watch-blog
watch-blog:
	./bin/watch-hindsite.sh $(BLOG_DIR) -content $(BLOG_DIR)/template/init

.PHONY: validate-blog
validate-blog:
	for f in $$(find $(BLOG_DIR)/build -name "*.html"); do echo $$f; html-validator --verbose --format=text --file=$$f; done