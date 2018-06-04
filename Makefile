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
