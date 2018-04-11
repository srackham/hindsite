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
	go install ./...

.PHONY: build
build:
	mkdir -p ./bin
	GOOS=linux GOARCH=amd64 go build -o ./bin/hindsite-linux-amd64 ./...
	GOOS=darwin GOARCH=amd64 go build -o ./bin/hindsite-darwin-amd64 ./...
	GOOS=windows GOARCH=amd64 go build -o ./bin/hindsite-windows-amd64.exe ./...
	GOOS=windows GOARCH=386 go build -o ./bin/hindsite-windows-386.exe ./...

.PHONY: test
test: bindata
	go test ./...

.PHONY: clean
clean:
	go clean -i ./...

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