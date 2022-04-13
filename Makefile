# hindsite Makefile

# Set defaults (see http://clarkgrubb.com/makefile-style-guide#prologue)
MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := test
.DELETE_ON_ERROR:
.SUFFIXES:
.ONESHELL:
 .SILENT:

GOFLAGS ?=
PACKAGES = ./fsutil ./site ./slice

.PHONY: install
install:
	LDFLAGS="-X main.BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)"
	# The version number is set to the tag of latest commit
	VERS="$$(git tag --points-at HEAD)"
	if [ -n "$$VERS" ]; then
		[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "illegal VERS=$$VERS " && exit 1
		LDFLAGS="$$LDFLAGS -X main.VERS=$$VERS"
	fi
	LDFLAGS="$$LDFLAGS -X main.OS=$$(go env GOOS)/$$(go env GOARCH)"
	go install -ldflags "$$LDFLAGS" ./...

.PHONY: test
test: install
	go vet $(PACKAGES)
	go test -cover $(PACKAGES)

.PHONY: clean
clean: fmt
	go mod verify
	go mod tidy
	go clean -i $(PACKAGES)

.PHONY: fmt
fmt:
	gofmt -w -s $$(find . -name '*.go')

.PHONY: tag
# Tag the latest commit with the VERS environment variable e.g. make tag VERS=v1.0.0
tag:
	[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "error: illegal VERS=$$VERS " && exit 1
	git tag -a -m "$$VERS" $$VERS

.PHONY: push
push: test validate-docs
	git push -u --tags origin master
	make submit-sitemap

DIST_DIR := ./dist

.PHONY: build-dist
# Build executable distributions and compress them to Zip files.
# Because the distribution is built from the working directory the working directory cannot contain
# uncommitted changes and the latest commit must be tagged with a release version number.
build-dist: clean test validate-docs
	[[ -n "$$(git status --porcelain)" ]] && echo "error: there are uncommitted changes in the working directory" && exit 1
	VERS="$$(git tag --points-at HEAD)"
	[[ -z "$$VERS" ]] && echo "error: the latest commit has not been tagged" && exit 1
	[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "error: illegal version tag: $$VERS " && exit 1
	[[ $$(ls $(DIST_DIR)/hindsite-$$VERS* 2>/dev/null | wc -w) -gt 0 ]] && echo "error: built version $$VERS already exists" && exit 1
	mkdir -p $(DIST_DIR)
	BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)
	COMMIT=$$(git rev-parse HEAD)
	BUILD_FLAGS="-X main.BUILT=$$BUILT -X main.COMMIT=$$COMMIT -X main.VERS=$$VERS"
	build () {
		export GOOS=$$1
		export GOARCH=$$2
		LDFLAGS="$$BUILD_FLAGS -X main.OS=$$GOOS/$$GOARCH"
		LDFLAGS="$$LDFLAGS -s -w"	# Strip symbols to decrease executable size
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
	cd $(DIST_DIR)
	build linux amd64
	build darwin amd64
	build windows amd64
	build windows 386
	sha1sum hindsite-$$VERS*.zip > hindsite-$$VERS-checksums-sha1.txt

.PHONY: release
# Upload release binary distributions for the version assigned to the VERS environment variable e.g. make release VERS=v1.0.0
release:
	REPO=hindsite
	USER=srackham
	[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "error: illegal VERS=$$VERS " && exit 1
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
	cd $(DIST_DIR)
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

# Generate build, serve and validate rules for builtin templates:
#
#	build-builtin-hello, server-builtin-hello, validate-builtin-hello
#	build-builtin-blog, server-builtin-blog, validate-builtin-blog
#	build-builtin-docs, server-builtin-docs, validate-builtin-docs

# Rule templates.
define rules_template
.PHONY: build-builtin-$(1)
build-builtin-$(1):
	hindsite build $(2) -content $(2)/template/init -lint -v

.PHONY: serve-builtin-$(1)
serve-builtin-$(1):
	hindsite serve $(2) -content $(2)/template/init -launch -navigate -lint -v

.PHONY: validate-builtin-$(1)
validate-builtin-$(1): build-builtin-$(1)
	for f in $$$$(find $(2)/build -name "*.html"); do
		# Skip page (it has custom Google CSE elements that fail validation).
		if [ "$$$$f" != "$(2)/build/search.html" ]; then
			echo $$$$f
			html-validator --verbose --format=text --file=$$$$f
		fi	
	done
endef

# Rule generation.
templates := hello blog docs
$(foreach t,$(templates),$(eval $(call rules_template,$(t),./cmd/hindsite/builtin/$(t))))


#
# Documentation tasks
#
HOMEPAGE = https://srackham.github.io/hindsite

.PHONY: build-docs
build-docs: install
	hindsite build docsite -build docs -lint
	make build-sitemap

.PHONY: serve-docs
serve-docs: install
	hindsite serve docsite -build docs -launch -navigate -lint -v

.PHONY: validate-docs
validate-docs: build-docs
	for f in $$(ls ./docs/{changelog,faq,index}.html); do echo $$f; html-validator --verbose --format text --file $$f; done

# Build Google search engine site map (see https://support.google.com/webmasters/answer/183668)
# index.html file URLs are converted to the canonical format with trailing slash character.
.PHONY: build-sitemap
build-sitemap:
	cd docs
	ls ./*.html \
	| grep -v './google' \
	| sed -e 's|^.|$(HOMEPAGE)|' \
	| sed -e 's|\/index.html$$|/|' \
	> sitemap.txt
	cd ..

# Submit site map to Google (see https://developers.google.com/search/docs/advanced/sitemaps/build-sitemap#addsitemap)
.PHONY: submit-sitemap
submit-sitemap:
	curl https://www.google.com/ping?sitemap=$(HOMEPAGE)/sitemap.txt
