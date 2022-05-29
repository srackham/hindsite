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
PACKAGES = ./...
XFLAG_PATH = github.com/srackham/hindsite/site

.PHONY: install
install:
	LDFLAGS="-X $(XFLAG_PATH).BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)"
	# The version number is set to the tag of latest commit
	VERS="$$(git tag --points-at HEAD)"
	if [ -n "$$VERS" ]; then
		[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "illegal VERS=$$VERS " && exit 1
		LDFLAGS="$$LDFLAGS -X $(XFLAG_PATH).VERS=$$VERS"
	fi
	LDFLAGS="$$LDFLAGS -X $(XFLAG_PATH).OS=$$(go env GOOS)/$$(go env GOARCH)"
	go install -ldflags "$$LDFLAGS"

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
push: test
	git push -u --tags origin master
	make --silent submit-sitemap

DIST_DIR := ./dist

.PHONY: build-dist
# Build executable distributions and compress them to Zip files.
# The VERS environment variable sets version number.
# If VERS is not set the version number defaults to v0.0.0 and version tag
# checks are skipped (v0.0.0 is reserved for testing only).
#
# Normally you want to build from a version-tagged commit, if it
# is not the current head then: stash current changes; temporarily checkout the
# tagged commit; run the build-dist task; revert to previous commit; pop the
# stash e.g. to make a distribution for version v.1.4.0:
#
#   git stash       	# Stash working directory changes
#   git checkout v1.4.0
#   make build-dist
#   git checkout master	# Restore previous commit
#   git stash pop   	# Restore previous working directory changes

# build-dist: clean test validate-docs
build-dist:
	VERS=$${VERS:-v0.0.0}	# v0.0.0 is for testing.
	[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "error: illegal version tag: $$VERS " && exit 1
	if [[ $$VERS != "v0.0.0" ]]; then
		[[ $$(ls $(DIST_DIR)/hindsite-$$VERS* 2>/dev/null | wc -w) -gt 0 ]] && echo "error: built version $$VERS already exists" && exit 1
		headtag="$$(git tag --points-at HEAD)"
		[[ -z "$$headtag" ]] && echo "error: the latest commit has not been tagged" && exit 1
		[[ $$headtag != $$VERS ]] && echo "error: the latest commit tag does not equal $$VERS" && exit 1
		[[ -n "$$(git status --porcelain)" ]] && echo "error: changes in the working directory" && exit 1
	else
		echo "WARNING: no VERS env variable specified, defaulting to v0.0.0 test build"
	fi
	mkdir -p $(DIST_DIR)
	BUILT=$$(date +%Y-%m-%dT%H:%M:%S%:z)
	COMMIT=$$(git rev-parse HEAD)
	BUILD_FLAGS="-X $(XFLAG_PATH).BUILT=$$BUILT -X $(XFLAG_PATH).COMMIT=$$COMMIT -X $(XFLAG_PATH).VERS=$$VERS"
	build () {
		export GOOS=$$1
		export GOARCH=$$2
		LDFLAGS="$$BUILD_FLAGS -X $(XFLAG_PATH).OS=$$GOOS/$$GOARCH"
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
		go build -ldflags "$$LDFLAGS" -o $$EXE ..
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
	REPO=srackham/hindsite
	[[ ! $$VERS =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] && echo "error: illegal VERS=$$VERS " && exit 1
	upload () {
		export GOOS=$$1
		export GOARCH=$$2
		FILE=hindsite-$$VERS-$$GOOS-$$GOARCH.zip
		gh release upload $$VERS --repo $$REPO $$FILE
	}
	gh release create $$VERS --repo $$REPO --draft --title "Hindsite $$VERS" --notes "Hindsite is a fast, lightweight static website generator."
	sleep 5	# Wait to avoid "release not found" error.
	cd $(DIST_DIR)
	upload linux amd64
	upload darwin amd64
	upload windows amd64
	upload windows 386
	SUMS=hindsite-$$VERS-checksums-sha1.txt
	gh release upload $$VERS --repo $$REPO $$SUMS

# Generate build, serve and validate rules for builtin templates:
#
#	build-builtin-hello, serve-builtin-hello, validate-builtin-hello
#	build-builtin-blog, serve-builtin-blog, validate-builtin-blog
#	build-builtin-docs, serve-builtin-docs, validate-builtin-docs

# Rule templates.
define rules_template
.PHONY: build-builtin-$(1)
build-builtin-$(1):
	hindsite build -site $(2) -content $(2)/template/init -lint -v

.PHONY: serve-builtin-$(1)
serve-builtin-$(1):
	hindsite serve -site $(2) -content $(2)/template/init -launch -navigate -lint -v

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
$(foreach t,$(templates),$(eval $(call rules_template,$(t),./site/builtin/$(t))))


#
# Documentation tasks
#
HOMEPAGE = https://srackham.github.io/hindsite

.PHONY: build-docs
build-docs: install
	rm -rf docs/*
	mkdir -p docs/builtin/blog
	hindsite build -site site/builtin/blog -content site/builtin/blog/template/init -build docs/builtin/blog -var urlprefix=/hindsite/builtin/blog -lint
	mkdir -p docs/builtin/docs
	hindsite build -site site/builtin/docs -content site/builtin/docs/template/init -build docs/builtin/docs -var urlprefix=/hindsite/builtin/docs -lint
	mkdir -p docs/builtin/hello
	hindsite build -site site/builtin/hello -content site/builtin/hello/template/init -build docs/builtin/hello -var urlprefix=/hindsite/builtin/hello -lint
	hindsite build -site docsite -build docs -keep -lint
	make --silent build-sitemap

.PHONY: serve-docs
serve-docs: install
	hindsite serve -site docsite -build docs -keep -launch -navigate -lint

.PHONY: validate-docs
validate-docs: build-docs
	for f in $$(find ./docs -name '*.html' -not -name 'google*.html' -not -name 'search.html'); do echo $$f; html-validator --verbose --format text --file $$f; done

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

#
# Build and validate docs and built-in templates.
#
.PHONY: validate-all
validate-all: validate-docs validate-builtin-hello validate-builtin-blog  validate-builtin-docs 

#
# Create and validate build file checksums for the testdata site.
#
TEST_SITE = ./site/testdata/blog
make-checksums:
	tmpdir=$$(mktemp -d /tmp/hindsite-XXXXXXXX)
	hindsite init -site $$tmpdir -from $(TEST_SITE)/template
	hindsite build -site $$tmpdir
	(cd $$tmpdir && find build -name '*.html' -type f -exec sha256sum '{}' +) > $(TEST_SITE)/checksums.txt

validate-checksums:
	tmpdir=$$(mktemp -d /tmp/hindsite-XXXXXXXX)
	hindsite init -site $$tmpdir -from $(TEST_SITE)/template
	hindsite build -site $$tmpdir
	f=$$(readlink -f $(TEST_SITE)/checksums.txt)
	(cd $$tmpdir && sha256sum --quiet --check $$f)

serve-testdata:
	tmpdir=$$(mktemp -d /tmp/hindsite-XXXXXXXX)
	hindsite init -site $$tmpdir -from $(TEST_SITE)/template
	hindsite serve -site $$tmpdir -launch -lint