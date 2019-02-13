#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
SOURCES := $(shell find . -name "*.go" | grep -v test.go | grep -v '\.\#*')
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	# TAGGED_VERSION := $(shell git describe --tags)
	# This doesn't work in CI, need to find another way...
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

LDFLAGS := "-X github.com/solo-io/sqoop/version.Version=$(VERSION)"
GCFLAGS := all="-N -l"

#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

.PHONY: update-deps
update-deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
	go get -u github.com/lyft/protoc-gen-validate
	go get -u github.com/paulvollmer/2gobytes

.PHONY: pin-repos
pin-repos:
	go run pin_repos.go

.PHONY: check-format
check-format:
	NOT_FORMATTED=$$(gofmt -l ./pkg/ ./test/) && if [ -n "$$NOT_FORMATTED" ]; then echo These files are not formatted: $$NOT_FORMATTED; exit 1; fi

check-spelling:
	./ci/spell.sh check


.PHONY: generated-code
generated-code: $(OUTPUT_DIR)/.generated-code

SUBDIRS:=pkg cmd
$(OUTPUT_DIR)/.generated-code:
	go generate ./...
	gofmt -w $(SUBDIRS)
	goimports -w $(SUBDIRS)
	mkdir -p $(OUTPUT_DIR)
	touch $@

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean:
	rm -rf _output
	rm -fr site




#----------------------------------------------------------------------------------
# sqoopctl
#----------------------------------------------------------------------------------

#.PHONY: sqoopctl
#sqoopctl: $(OUTPUT_DIR)/sqoopctl
#
#$(OUTPUT_DIR)/sqoopctl: $(SOURCES)
#	go build -v -o $@ cmd/sqoopctl/main.go
#
#$(OUTPUT_DIR)/sqoopctl-linux-amd64: $(SOURCES)
#	GOOS=linux go build -v -o $@ cmd/sqoopctl/main.go
#
#$(OUTPUT_DIR)/sqoopctl-darwin-amd64: $(SOURCES)
#	GOOS=darwin go build -v -o $@ cmd/sqoopctl/main.go

#----------------------------------------------------------------------------------
# Docs
#----------------------------------------------------------------------------------

docs/api.json: $(PROTOS)
	export DISABLE_SORT=1 && \
	cd api/v1/ && \
	mkdir -p $(ROOTDIR)/pkg/api/types/v1 && \
	protoc \
	-I=. \
	-I=$(GOPATH)/src \
	-I=$(GOPATH)/src/github.com/gogo/protobuf/ \
	--plugin=protoc-gen-doc=$(GOPATH)/bin/protoc-gen-doc \
    --doc_out=$(ROOTDIR)/docs/ \
    --doc_opt=json,api.json \
	./*.proto

docs/index.md: README.md
	cat README.md | sed 's@docs/@@' > docs/index.md

docs/getting_started/kubernetes/1.md: examples/petstore/README.md
	mkdir -p docs/getting_started/kubernetes/
	cp examples/petstore/README.md $@

doc: docs/api.json docs/index.md docs/getting_started/kubernetes/1.md
	go run docs/gen_docs.go

site: doc
	mkdocs build

deploy-site: site
	firebase deploy --only hosting:sqoop-site

#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------

RELEASE_BINARIES := $(OUTPUT_DIR)/sqoopctl-linux-amd64 $(OUTPUT_DIR)/sqoopctl-darwin-amd64

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: release
release: release-binaries
	hack/create-release.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=sqoop tag=v$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),hack/upload-github-release-asset.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=sqoop tag=v$(VERSION) filename=$(BINARY);)
