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
CMD := main
LDFLAGS := "-X '${CMD}.Branch=${BRANCH}' -X '${CMD}.Timestamp=${BUILD_DATE}' -X '${CMD}.Revision=${HASH}'"
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

build: sqlc-generate
	$(Q) $(GOBUILD) -o $(BINARY_NAME) -ldflags=$(LDFLAGS)

test:
	$(Q) $(GOTEST) ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_LINUX)

run:
	$(Q) $(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

# Cross compilation
build-linux: sqlc-generate
	$(Q) CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_LINUX) -ldflags=$(LDFLAGS)

#docker-build:
#	docker run --rm -it -v "$(GOPATH)":/go -w /go/src/bitbucket.org/rsohlich/makepost golang:latest go build -o "$(BINARY_LINUX)" -v

docker-build:
	$(Q) DOCKER_BUILDKIT=1 docker build --secret id=token,src=${BITBUCKET_TOKEN} --tag ${DOCKER_TAG}  --build-arg proj_name=${PROJ_NAME} --build-arg ldflags=${LDFLAGS} -f build/package/Dockerfile .

docker-push-flexid-dev:
	$(Q) docker tag ${DOCKER_TAG} ${DEV_ECR_HOST}/${DOCKER_TAG}
	$(Q) AWS_PROFILE=development-flexid aws ecr get-login-password --region ${DEV_REGION} | docker login --username AWS --password-stdin ${DEV_ECR_HOST}
	$(Q) AWS_PROFILE=development-flexid docker push ${DEV_ECR_HOST}/${DOCKER_TAG}
