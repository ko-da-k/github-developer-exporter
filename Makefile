.PHONY: all run build build-linux test clean fmt docker/* gcloud/*

VERSION := $(shell git describe --tags --abbrev=0 --always)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.revision=$(REVISION)'
GOCMD := GO111MODULE=on go

BINARY := app
TARGET := main.go

DOCKER_IMAGE := github-developer-exporter:0.1.0

all: build

run:
	$(GOCMD) run $(TARGET)

build:
	$(GOCMD) build --ldflags="$(LDFLAGS)" -o $(BINARY) $(TARGET)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOCMD) build -o $(BINARY) $(TARGET)

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run -p 8888:8888 -e PORT=$(PORT) $(DOCKER_IMAGE)

compose:
	docker-compose -f ./local/docker-compose.yml up

compose-rm:
	docker-compose -f ./local/docker-compose.yml rm

test:
	# without infrastructure layer
	$(GOCMD) test -v -bench=. ./interfaces/handlers ./domain/... -benchmem

test-all:
	$(GOCMD) test -v -bench=. ./... -benchmem

coverage:
	$(GOCMD) test -coverprofile=profile ./... && $(GOCMD) tool cover -html=profile

cobertura:
	$(GOCMD) test -coverprofile=profile ./interfaces/handlers ./domain/... ./config && gocover-cobertura < profile > coverage.xml

clean:
	$(GOCMD) clean
	rm -f $(BINARY)

fmt:
	$(GOCMD) fmt ./...
