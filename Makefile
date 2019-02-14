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
# sqoop
#----------------------------------------------------------------------------------

SQOOP_DIR=.
SQOOP_SOURCES=$(call get_sources,$(SQOOP_DIR))

$(OUTPUT_DIR)/sqoop-linux-amd64: $(SQOOP_SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(SQOOP_DIR)/cmd/main.go


.PHONY: sqoop
sqoop: $(OUTPUT_DIR)/sqoop-linux-amd64

$(OUTPUT_DIR)/Dockerfile.sqoop: $(SQOOP_DIR)/cmd/Dockerfile
	cp $< $@

sqoop-docker: $(OUTPUT_DIR)/sqoop-linux-amd64 $(OUTPUT_DIR)/Dockerfile.sqoop
	docker build -t soloio/sqoop:$(VERSION)  $(OUTPUT_DIR) -f $(OUTPUT_DIR)/Dockerfile.sqoop

#----------------------------------------------------------------------------------
# Deployment Manifests / Helm
#----------------------------------------------------------------------------------


HELM_SYNC_DIR := $(OUTPUT_DIR)/helm
HELM_DIR := install/helm
MANIFEST_DIR := install/manifest

.PHONY: manifest
manifest: helm-template init-helm install/manifest/sqoop.yaml update-helm-chart

# creates Chart.yaml, values.yaml, and requirements.yaml
.PHONY: helm-template
helm-template:
	mkdir -p $(MANIFEST_DIR)
	go run install/helm/sqoop/generate.go $(VERSION)

update-helm-chart:
ifeq ($(RELEASE),"true")
	mkdir -p $(HELM_SYNC_DIR)/charts
	helm package --destination $(HELM_SYNC_DIR)/charts $(HELM_DIR)/sqoop
	helm repo index $(HELM_SYNC_DIR)
endif

install/manifest/sqoop.yaml: helm-template
	helm template install/helm/sqoop --namespace sqoop --name=sqoop > $@

init-helm:
	helm repo add gloo https://storage.googleapis.com/solo-public-helm
	helm dependency update install/helm/sqoop

#----------------------------------------------------------------------------------
# sqoopctl
#----------------------------------------------------------------------------------

CLI_DIR=cli

$(OUTPUT_DIR)/sqoopctl: $(SOURCES)
	go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go


$(OUTPUT_DIR)/sqoopctl-linux-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go


$(OUTPUT_DIR)/sqoopctl-darwin-amd64: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go

$(OUTPUT_DIR)/sqoopctl-windows-amd64.exe: $(SOURCES)
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags=$(LDFLAGS) -gcflags=$(GCFLAGS) -o $@ $(CLI_DIR)/cmd/main.go


.PHONY: sqoopctl
sqoopctl: $(OUTPUT_DIR)/sqoopctl
.PHONY: sqoopctl-linux-amd64
sqoopctl-linux-amd64: $(OUTPUT_DIR)/sqoopctl-linux-amd64
.PHONY: sqoopctl-darwin-amd64
sqoopctl-darwin-amd64: $(OUTPUT_DIR)/sqoopctl-darwin-amd64
.PHONY: sqoopctl-windows-amd64
sqoopctl-windows-amd64: $(OUTPUT_DIR)/sqoopctl-windows-amd64.exe

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
GH_ORG:=solo-io
GH_REPO:=sqoop

# For now, expecting people using the release to start from a sqoopctl CLI we provide, not
# installing the binaries locally / directly. So only uploading the CLI binaries to Github.
# The other binaries can be built manually and used, and docker images for everything will
# be published on release.
RELEASE_BINARIES :=
ifeq ($(RELEASE),"true")
	RELEASE_BINARIES := \
		$(OUTPUT_DIR)/sqoopctl-linux-amd64 \
		$(OUTPUT_DIR)/sqoopctl-darwin-amd64 \
		$(OUTPUT_DIR)/sqoopctl-windows-amd64.exe
endif

RELEASE_YAMLS :=
ifeq ($(RELEASE),"true")
	RELEASE_YAMLS := \
		install/manifest/sqoop.yaml
endif

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: release-yamls
release-yamls: $(RELEASE_YAMLS)

# This is invoked by cloudbuild. When the bot gets a release notification, it kicks of a build with and provides a tag
# variable that gets passed through to here as $TAGGED_VERSION. If no tag is provided, this is a no-op. If a tagged
# version is provided, all the release binaries are uploaded to github.
# Create new releases by clicking "Draft a new release" from https://github.com/solo-io/sqoop/releases
.PHONY: release
release: release-binaries release-yamls
ifeq ($(RELEASE),"true")
	@$(foreach BINARY,$(RELEASE_BINARIES),ci/upload-github-release-asset.sh owner=solo-io repo=sqoop tag=$(TAGGED_VERSION) filename=$(BINARY) sha=TRUE;)
	@$(foreach YAML,$(RELEASE_YAMLS),ci/upload-github-release-asset.sh owner=solo-io repo=sqoop tag=$(TAGGED_VERSION) filename=$(YAML);)
endif

#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
#
#---------
#--------- Push
#---------

DOCKER_IMAGES :=
ifeq ($(RELEASE),"true")
	DOCKER_IMAGES := docker
endif

.PHONY: docker docker-push
docker: sqoop-docker

# Depends on DOCKER_IMAGES, which is set to docker if RELEASE is "true", otherwise empty (making this a no-op).
# This prevents executing the dependent targets if RELEASE is not true, while still enabling `make docker`
# to be used for local testing.
# docker-push is intended to be run by CI
docker-push: $(DOCKER_IMAGES)
ifeq ($(RELEASE),"true")
	docker push soloio/sqoop:$(VERSION)
endif

