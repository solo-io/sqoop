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
	go run ci/pin_repos.go

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
	(rm docs/cli/sqoopctl* && go run cli/cmd/docs/main.go)
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
	helm template install/helm/sqoop --namespace gloo-system --name=sqoop > $@

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
# Release
#----------------------------------------------------------------------------------

.PHONY: upload-github-release-assets
upload-github-release-assets:
	go run ci/upload_github_release_assets.go

.PHONY: push-docs
push-docs:
	go run ci/push_docs.go

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

