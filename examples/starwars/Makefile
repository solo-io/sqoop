SOURCES := $(shell find . -name "*.go")

BINARY:=starwars
IMAGE_TAG?=latest
build: $(BINARY)
OUTPUT_DIR ?= .

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 GOOS=linux go build -v -i -o $(OUTPUT_DIR)/$@ *.go

docker: $(BINARY)
	docker build -t soloio/$(BINARY)-example:$(IMAGE_TAG) .

push: docker
	docker push soloio/$(BINARY)-example:$(IMAGE_TAG)

clean:
	rm -f $(BINARY)
