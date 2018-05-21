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

bindata.go: $(BINDATA_FILES)
	go-bindata -prefix ./builtin/ -ignore '/(build|content)/' ./builtin/...

.PHONY: bindata
bindata: bindata.go

.PHONY: install
install: test
	go install ./...

.PHONY: test
test: bindata
	go test ./...

.PHONY: clean
clean:
	go clean -i ./...

.PHONY: push
push:
	git push -u --tags origin master

.PHONY: build-docs
build-docs: install
	hindsite build docsrc -build docs

.PHONY: serve-docs
serve-docs: install
	hindsite serve docsrc -build docs -launch -v

.PHONY: validate-docs
validate-docs: build-docs
	for f in $$(find ./docs -name "*.html"); do echo $$f; html-validator --verbose --format text --file $$f; done

.PHONY: build
build: build-docs
	mkdir -p ./bin
	BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)
	COMMIT=$$(git rev-parse HEAD)
	VERS=$$(git describe --tags --abbrev=0)
	BUILD_FLAGS="-X main.BUILT=$$BUILT -X main.COMMIT=$$COMMIT -X main.VERS=$$VERS"
	build () {
		export GOOS=$$1
		export GOARCH=$$2
		LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
		NAME=hindsite-$$VERS-$$GOOS-$$GOARCH
		EXE=$$NAME/hindsite
		if [ "$$1" = "windows" ]; then
			EXE=$$EXE.exe
		fi
		ZIP=$$NAME.zip
		rm -f $$ZIP
		rm -rf $$NAME
		mkdir $$NAME
		cp ../LICENSE $$NAME
		cp ../README.md $$NAME/README.txt
		go build -ldflags "$$LDFLAGS" -o $$EXE ../...
		zip $$ZIP $$NAME/*
	}
	cd bin
	build linux amd64
	build darwin amd64
	build windows amd64
	build windows 386
	sha1sum hindsite-$$VERS*.zip > hindsite-$$VERS-checksums-sha1.txt

.PHONY: release
release:
	REPO=hindsite
	USER=srackham
	VERS=$$(git describe --tags --abbrev=0)
	upload () {
		export GOOS=$$1
		export GOARCH=$$2
		FILE=hindsite-$$VERS-$$GOOS-$$GOARCH.zip
		github-release upload \
			--user $$USER \
			--repo $$REPO \
			--tag $$VERS \
			--name $$FILE \
			--file $$FILE
	}
	github-release release \
		--user $$USER \
		--repo $$REPO \
		--tag $$VERS \
		--name "hindsite $$VERS" \
		--description "hindsite is a fast, lightweight static website generator."
	cd bin
	upload linux amd64
	upload darwin amd64
	upload windows amd64
	upload windows 386
	SUMS=hindsite-$$VERS-checksums-sha1.txt
	github-release upload \
		--user $$USER \
		--repo $$REPO \
		--tag $$VERS \
		--name $$SUMS \
		--file $$SUMS

BLOG_DIR = ./builtin/blog

# Built the builtin blog's init directory.
.PHONY: build-blog
build-blog: install
	hindsite build $(BLOG_DIR) -content $(BLOG_DIR)/template/init -v

.PHONY: serve-blog
serve-blog: install
	hindsite serve $(BLOG_DIR) -content $(BLOG_DIR)/template/init -launch -v

.PHONY: validate-blog
validate-blog: build-blog
	for f in $$(find $(BLOG_DIR)/build -name "*.html"); do echo $$f; html-validator --verbose --format text --file $$f; done