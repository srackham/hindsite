# hindsite Makefile

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
	BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)
	COMMIT=$$(git rev-parse HEAD)
	VERS=$$(git describe --tags --abbrev=0)
	BUILD_FLAGS="-X main.BUILT=$$BUILT -X main.COMMIT=$$COMMIT -X main.VERS=$$VERS"

	cd bin
	cp ../LICENSE .
	cp ../README.md README

	export GOOS=linux
	export GOARCH=amd64
	LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
	EXE=hindsite-$$VERS-$$GOOS-$$GOARCH
	ZIP=$$EXE.zip
	go build -ldflags "$$LDFLAGS" -o $$EXE ../...
	zip $$ZIP README LICENSE $$EXE

	export GOOS=darwin
	export GOARCH=amd64
	LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
	EXE=hindsite-$$VERS-$$GOOS-$$GOARCH
	ZIP=$$EXE.zip
	go build -ldflags "$$LDFLAGS" -o $$EXE ../...
	zip $$ZIP README LICENSE $$EXE

	export GOOS=windows
	export GOARCH=amd64
	LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
	EXE=hindsite-$$VERS-$$GOOS-$$GOARCH
	ZIP=$$EXE.zip
	go build -ldflags "$$LDFLAGS" -o $$EXE.exe ../...
	zip $$ZIP README LICENSE $$EXE.exe

	export GOOS=windows
	export GOARCH=386
	LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
	EXE=hindsite-$$VERS-$$GOOS-$$GOARCH
	ZIP=$$EXE.zip
	go build -ldflags "$$LDFLAGS" -o $$EXE.exe ../...
	zip $$ZIP README LICENSE $$EXE.exe

	sha1sum hindsite-*.zip > SHA1SUM
	md5sum hindsite-*.zip > MD5SUM

.PHONY: test
test: bindata
	go test ./...

.PHONY: clean
clean:
	go clean -i ./...

.PHONY: push
push:
	git push -u --tags origin master

.PHONY: build-doc
build-doc: install
	hindsite build doc

.PHONY: serve-doc
serve-doc: install
	hindsite serve doc

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
serve-blog: install
	hindsite serve $(BLOG_DIR) -content $(BLOG_DIR)/template/init

.PHONY: validate-blog
validate-blog: build-blog
	for f in $$(find $(BLOG_DIR)/build -name "*.html"); do echo $$f; html-validator --verbose --format text --file $$f; done