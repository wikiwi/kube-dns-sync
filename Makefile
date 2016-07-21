# Copyright (C) 2016 wikiwi.io
#
# This software may be modified and distributed under the terms
# of the MIT license. See the LICENSE file for details.

###  Configuration ###
GO_PACKAGE     ?= github.com/wikiwi/kube-dns-sync
REPOSITORY     ?= wikiwi/kube-dns-sync

### Docker Tag settings ###
LATEST_VERSION := 0.1

### Coverage settings ###
COVER_PACKAGES = $(shell cd pkg && go list -f '{{.ImportPath}}' ./... | tr '\n' ',' | sed 's/.$$//')

### Build Tools ###
GO ?= go
GLIDE ?= glide
GIT ?= git
DOCKER ?= docker
GOVER ?= gover

# Glide Options
GLIDE_OPTS ?=
GLIDE_GLOBAL_OPTS ?=

### CI Settings ###
# Set branch with most current HEAD of master e.g. master or origin/master.
# E.g. Gitlab doesn't pull the master branch but fetches it to origin/master.
MASTER_BRANCH ?= master

### Environment ###
HAS_GLIDE := $(shell command -v ${GLIDE};)
HAS_GIT := $(shell command -v ${GIT};)
HAS_GO := $(shell command -v ${GO};)
GOOS := $(shell ${GO} env GOOS)
GOARCH := $(shell ${GO} env GOARCH)
BINARIES := $(notdir $(wildcard cmd/*))

# Load versioning logic.
include Makefile.versioning

# Docker Image info.
IMAGE := ${REPOSITORY}:${BUILD_REF}

# Show build info.
info:
	@echo "Version: ${BUILD_VERSION}"
	@echo "Image:   ${IMAGE}"
	@echo "Tags:    ${TAGS}"

.PHONY: build
ifneq (${GOOS}, "windows")
build: ${BINARIES:%=bin/${GOOS}/${GOARCH}/%}
else
build: ${BINARIES:%=bin/${GOOS}/${GOARCH}/%.exe}
endif

.PHONY: build-cross
build-cross: ${BINARIES:%=build-cross-%}
build-cross-%: bin/linux/amd64/% bin/freebsd/amd64/% bin/darwin/amd64/% bin/windows/amd64/%.exe
	$(NOOP)

.PHONY: build-for-docker
build-for-docker: ${BINARIES:%= bin/linux/amd64/%}

# docker-build will build the docker image.
.PHONY: docker-build
docker-build: build-for-docker
	${DOCKER} build --pull -t ${IMAGE} .

# docker-push will push all tags to the repository
.PHONY: docker-push
docker-push: ${TAGS:%=docker-push-%}
docker-push-%:
	${DOCKER} tag ${IMAGE} ${REPOSITORY}:$* && docker push ${REPOSITORY}:$*

.PHONY: has-tags
has-tags:
ifndef TAGS
	@echo No tags set for this build
	false
endif

# clean deletes build artifacts from the project.
.PHONY: clean
clean:
	rm -rf bin artifacts

# test will start the project test suites.
.PHONY: test
test:
	echo Running unit tests
	cd cmd && go test ./...
	cd pkg && go test ./...
	echo Running integration tests
	cd test && go test ./...

.PHONY: test-with-coverage
test-with-coverage:
	cd cmd && go test ./...
	cd pkg && go list -f "{{if len .TestGoFiles}}go test -coverpkg=\"${COVER_PACKAGES}\" -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}};{{end}}" ./... | sh
	cd test && go list -f "{{if len .TestGoFiles}}go test -coverpkg=\"${COVER_PACKAGES}\" -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}};{{end}}" ./... | sh
	${GOVER}

# bootstrap will install project dependencies.
.PHONY: bootstrap
bootstrap:
ifndef HAS_GO
	$(error You must install Go)
endif
ifndef HAS_GIT
	$(error You must install Git)
endif
ifndef HAS_GLIDE
	${GO} get -u github.com/Masterminds/glide
endif
	${GLIDE} ${GLIDE_GLOBAL_OPTS} install ${GLIDE_OPTS}

include Makefile.build

