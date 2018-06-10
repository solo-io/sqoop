# Change this if your googleapis is in a different directory
export GOOGLE_PROTOS_HOME=$(HOME)/workspace/googleapis

ROOTDIR := $(shell pwd)
PROTOS := $(shell find api/v1 -name "*.proto")
SOURCES := $(shell find . -name "*.go" | grep -v test)
GENERATED_PROTO_FILES := $(shell find pkg/api/types/v1 -name "*.pb.go")
OUTPUT_DIR ?= _output

PACKAGE_PATH:=github.com/solo-io/qloo

#----------------------------------------------------------------------------------
# Build
#----------------------------------------------------------------------------------

# Generated code

.PHONY: all
all: build

.PHONY: proto
proto: $(GENERATED_PROTO_FILES)

$(GENERATED_PROTO_FILES): $(PROTOS)
	cd api/v1/ && \
	mkdir -p $(ROOTDIR)/pkg/api/types/v1 && \
	protoc \
	-I=. \
	-I=$(GOPATH)/src \
	-I=$(GOPATH)/src/github.com/gogo/protobuf/ \
	--gogo_out=$(GOPATH)/src \
	./*.proto

$(OUTPUT_DIR):
	mkdir -p $@

# kubernetes custom clientsets
.PHONY: clientset
clientset: $(GENERATED_PROTO_FILES) $(SOURCES)
	cd ${GOPATH}/src/k8s.io/code-generator && \
	./generate-groups.sh all \
		$(PACKAGE_PATH)/pkg/storage/crd/client \
		$(PACKAGE_PATH)/pkg/storage/crd \
		"solo.io:v1"

.PHONY: generated-code
generated-code:
	go generate ./...

$(OUTPUT):
	mkdir -p $(OUTPUT)

# Core Binaries

BINARIES ?= qloo
DEBUG_BINARIES = $(foreach BINARY,$(BINARIES),$(BINARY)-debug)

DOCKER_ORG=soloio

.PHONY: build
build: $(BINARIES)

.PHONY: debug-build
debug-build: $(DEBUG_BINARIES)

docker: $(foreach BINARY,$(BINARIES),$(shell echo $(BINARY)-docker))
docker-push: $(foreach BINARY,$(BINARIES),$(shell echo $(BINARY)-docker-push))

define BINARY_TARGETS
$(eval VERSION := $(shell cat version))
$(eval IMAGE_TAG ?= $(VERSION))
$(eval OUTPUT_BINARY := $(OUTPUT_DIR)/$(BINARY))

.PHONY: $(BINARY)
.PHONY: $(BINARY)-debug
.PHONY: $(BINARY)-docker
.PHONY: $(BINARY)-docker-debug
.PHONY: $(BINARY)-docker-push
.PHONY: $(BINARY)-docker-push-debug

# nice targets for the binaries
$(BINARY): $(OUTPUT_BINARY)
$(BINARY)-debug: $(OUTPUT_BINARY)-debug

# go build
$(OUTPUT_BINARY): $(OUTPUT_DIR) $(PREREQUISITES)
	CGO_ENABLED=0 GOOS=linux go build -v -o $(OUTPUT_BINARY) cmd/$(BINARY)/main.go
$(OUTPUT_BINARY)-debug: $(OUTPUT_DIR) $(PREREQUISITES)
	go build -i -gcflags "all=-N -l" -o $(OUTPUT_BINARY)-debug cmd/$(BINARY)/main.go

# docker
$(BINARY)-docker: $(OUTPUT_BINARY)
	docker build -t $(DOCKER_ORG)/$(BINARY):$(IMAGE_TAG) $(OUTPUT_DIR) -f - < cmd/$(BINARY)/Dockerfile
$(BINARY)-docker-debug: $(OUTPUT_BINARY)-debug
	docker build -t $(DOCKER_ORG)/$(BINARY)-debug:$(IMAGE_TAG) $(OUTPUT_DIR) -f - < cmd/$(BINARY)/Dockerfile.debug
$(BINARY)-docker-push: $(BINARY)-docker
	docker push $(DOCKER_ORG)/$(BINARY):$(IMAGE_TAG)
$(BINARY)-docker-push-debug: $(BINARY)-docker-debug
	docker push $(DOCKER_ORG)/$(BINARY)-debug:$(IMAGE_TAG)

endef

PREREQUISITES := $(SOURCES) $(GENERATED_PROTO_FILES) generated-code clientset
$(foreach BINARY,$(BINARIES),$(eval $(BINARY_TARGETS)))

# clean

clean:
	rm -rf $(OUTPUT_DIR)

#----------------------------------------------------------------------------------
# qlooctl
#----------------------------------------------------------------------------------

.PHONY: qlooctl
qlooctl: $(OUTPUT_DIR)/qlooctl

$(OUTPUT_DIR)/qlooctl: $(SOURCES)
	go build -v -o $@ cmd/qlooctl/main.go

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

doc: docs/api.json docs/index.md
	go run docs/gen_docs.go

site: doc
	mkdocs build

docker-docs: site
	docker build -t $(DOCKER_ORG)/qloo-docs:$(VERSION) -f Dockerfile.site .

