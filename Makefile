# hindsite Makefile

# Set defaults (see http://clarkgrubb.com/makefile-style-guide#prologue)
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:
.ONESHELL:
# .SILENT:

GOFLAGS ?=

.PHONY: install
install:
	# Set VERS environment variable to set the executable version number displayed by the `hindsite help` command..
	LDFLAGS="-X main.BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)"
	VERS=$${VERS:=}
	if [ -n "$$VERS" ]; then
		LDFLAGS="$$LDFLAGS -X main.VERS=$$VERS"
	fi
	LDFLAGS="$$LDFLAGS -X main.OS=$$(go env GOOS)/$$(go env GOARCH)"
	go install -ldflags "$$LDFLAGS" ./...

.PHONY: test
test: install
	go test -cover ./...

.PHONY: clean
clean:
	go mod verify
	go mod tidy
	go clean -i ./...

.PHONY: fmt
fmt:
	gofmt -w -s $$(find . -name '*.go')

.PHONY: push
push: test
	git push -u --tags origin master

.PHONY: build-docs
build-docs: install
	hindsite build docs
	cp docs/build/* docs	# Github pages serves from the ./docs folder

.PHONY: serve-docs
serve-docs: install
	hindsite serve docs -launch -navigate -v

.PHONY: validate-docs
validate-docs: build-docs
	for f in $$(ls ./docs/*.html); do echo $$f; html-validator --verbose --format text --file $$f; done

.PHONY: build-dist
build-dist: build-docs
	# Set VERS environment variable to override default version (the latest tag value) e.g. VERS=v1.0.0 make build
	mkdir -p ./bin
	BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)
	COMMIT=$$(git rev-parse HEAD)
	VERS=$${VERS:=}
	if [ -z "$$VERS" ]; then
		VERS=$$(git describe --tags --abbrev=0)
		if [ -z "$$VERS" ]; then
			echo "missing version tag"
			exit 1
		fi
	fi
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
	if [ $$(ls hindsite-$$VERS* 2>/dev/null | wc -w) -gt 0 ]; then
		echo "built version $$VERS already exists"
		exit 1
	fi
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

BLOG_DIR = ./cmd/hindsite/builtin/blog

# Built the builtin blog's init directory.
.PHONY: build-blog
build-blog: install
	hindsite build $(BLOG_DIR) -content $(BLOG_DIR)/template/init -v

.PHONY: serve-blog
serve-blog: install
	hindsite serve $(BLOG_DIR) -content $(BLOG_DIR)/template/init -launch -navigate -v

.PHONY: validate-blog
validate-blog: build-blog
	for f in $$(find $(BLOG_DIR)/build -name "*.html"); do
		# Skip page (it has custom Google CSE elements that fail validation).
		if [ "$$f" != "$(BLOG_DIR)/build/search.html" ]; then
			echo $$f
			html-validator --verbose --format=text --file=$$f
		fi	
	done