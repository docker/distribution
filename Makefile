# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Used to populate version variable in main package.
VERSION=$(shell git describe --match 'v[0-9]*' --dirty='.m' --always)
GO_LDFLAGS=-ldflags "-X `go list ./version`.Version $(VERSION)"

.PHONY: clean all fmt vet lint deps build test binaries
.DEFAULT: default
all: AUTHORS clean fmt vet fmt lint deps build test binaries

AUTHORS: .mailmap .git/HEAD
	 git log --format='%aN <%aE>' | sort -fu > $@

# This only needs to be generated by hand when cutting full releases.
version/version.go:
	./version/version.sh > $@

${PREFIX}/bin/registry: version/version.go $(shell find . -type f -name '*.go')
	@echo "+ $@"
	@godep go build -o $@ ${GO_LDFLAGS} ./cmd/registry

${PREFIX}/bin/registry-api-descriptor-template: version/version.go $(shell find . -type f -name '*.go')
	@echo "+ $@"
	@godep go build -o $@ ${GO_LDFLAGS} ./cmd/registry-api-descriptor-template

${PREFIX}/bin/dist: version/version.go $(shell find . -type f -name '*.go')
	@echo "+ $@"
	@godep go build -o $@ ${GO_LDFLAGS} ./cmd/dist

docs/spec/api.md: docs/spec/api.md.tmpl ${PREFIX}/bin/registry-api-descriptor-template
	./bin/registry-api-descriptor-template $< > $@

vet:
	@echo "+ $@"
	@godep go vet ./...

fmt:
	@echo "+ $@"
	@test -z "$$(GOPATH=`godep path` gofmt -s -l . | grep -v Godeps/_workspace/src/ | tee /dev/stderr)" || \
		echo "+ please format Go code with 'gofmt -s'"

lint:
	@echo "+ $@"
	@test -z "$$(GOPATH=`godep path` golint ./... | grep -v Godeps/_workspace/src/ | tee /dev/stderr)"

deps:
	@echo "+ $@"
	@godep restore

build:
	@echo "+ $@"
	@godep go build -v ${GO_LDFLAGS} ./...

test:
	@echo "+ $@"
	@godep go test -test.short ./...

test-full:
	@echo "+ $@"
	@godep go test ./...

binaries: ${PREFIX}/bin/registry ${PREFIX}/bin/registry-api-descriptor-template ${PREFIX}/bin/dist
	@echo "+ $@"

clean:
	@echo "+ $@"
	@rm -rf "${PREFIX}/bin/registry" "${PREFIX}/bin/registry-api-descriptor-template"

	
# Use the existing docs build cmds from docker/docker
# Later, we will move this into an import
DOCS_MOUNT := $(if $(DOCSDIR),-v $(CURDIR)/$(DOCSDIR):/$(DOCSDIR))
DOCSPORT := 8000
DOCKER_DOCS_IMAGE := docker-docs-$(VERSION)
DOCKER_RUN_DOCS := docker run --rm -it $(DOCS_MOUNT) -e AWS_S3_BUCKET -e NOCACHE

docs: docs-build
	$(DOCKER_RUN_DOCS) -p $(DOCSPORT):8000 "$(DOCKER_DOCS_IMAGE)" mkdocs serve

docs-shell: docs-build
	$(DOCKER_RUN_DOCS) -p $(DOCSPORT):8000 "$(DOCKER_DOCS_IMAGE)" bash

docs-build:
	docker build -t "$(DOCKER_DOCS_IMAGE)" -f docs/Dockerfile .
