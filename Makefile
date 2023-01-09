PROJ_NAME=bekind

 # Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=$(PROJ_NAME)
BINARY_LINUX=$(BINARY_NAME)_linux
HASH := $(shell git rev-parse --short HEAD)
COMMIT_DATE := $(shell git show -s --format=%ci ${HASH})
BUILD_DATE := $(shell date '+%Y-%m-%dT%H:%M:%S')
BRANCH := $(shell git branch --show-current)
PVARS := main
LDFLAGS := "-X '${PVARS}.Branch=${BRANCH}' -X '${PVARS}.Timestamp=${BUILD_DATE}' -X '${PVARS}.Revision=${HASH}'"
LINUX_FLAGS := CGO_ENABLED=0 GOOS=linux GOARCH=amd64
DOCKER_TAG := bekind:${HASH}

ifeq ($(VERBOSE),1)
  quiet =
  Q = DOCKER_BUILDKIT=0
else
  quiet=quiet_
  Q = @
endif

.PHONY: clean test build build-docker all

all: build

build:
	$(Q) cd cmd; $(GOBUILD) -o $(BINARY_NAME)  -ldflags=$(LDFLAGS) .

# Cross compilation
build-linux:
	$(Q) cd cmd; $(LINUX_FLAGS) $(GOBUILD) -o $(BINARY_LINUX)  -ldflags=$(LDFLAGS) .

test:
	$(Q) $(GOTEST) ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)

deploy-binary:
	$(Q) ansible-playbook -i deploy/hosts deploy/copy-binary.yml

deploy-web:
	$(Q) ansible-playbook -i deploy/hosts deploy/copy-web.yml
