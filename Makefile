VERSION := $(shell git describe --tags --dirty=-modified --always)
SMOKE_TEST_IMAGE := quay.io/skupper/skupper-ocp-smoke-test-image
DOCKER := docker
TEST_BINARIES_FOLDER := /image/bin
PLATFORM := linux/amd64,linux/s390x

all: build-smoke-image
	$(info ************  Information ************)
	$(info **** make all will only build the docker image.)
	$(info **** use make build-and-push if you want to build the image and push it to the repo)
	$(info **** you can also use make push-smoke-image to push an created image to the repo)

build-and-push: build-smoke-image push-smoke-image

build-smoke-image:
	$(info ************ Building the image ************)
	${DOCKER} buildx build -o type=docker --provenance false --sbom false --platform $(PLATFORM) -t ${SMOKE_TEST_IMAGE} .

push-smoke-image:
	$(info Tagging the image)
	${DOCKER} tag ${SMOKE_TEST_IMAGE} ${SMOKE_TEST_IMAGE}:${VERSION}
	$(info Pushing the image)
	${DOCKER} push ${SMOKE_TEST_IMAGE}
	${DOCKER} push ${SMOKE_TEST_IMAGE}:${VERSION}

build-tests:
	mkdir -p ${TEST_BINARIES_FOLDER}
	go test -c -v ./cmd/... -o ${TEST_BINARIES_FOLDER}/skupper-ocp-smoke-test
