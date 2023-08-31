VERSION := $(shell git describe --tags --dirty=-modified --always)
SMOKE_TEST_IMAGE := quay.io/skupper/skupper-ocp-smoke-test-image
DOCKER := docker

all: build-smoke-image push-smoke-image

build-smoke-image:
	${DOCKER} build --no-cache -t ${SMOKE_TEST_IMAGE} -f Dockerfile .

push-smoke-image:
	${DOCKER} tag ${SMOKE_TEST_IMAGE} ${SMOKE_TEST_IMAGE}:${VERSION}
	${DOCKER} push ${SMOKE_TEST_IMAGE}
	${DOCKER} push ${SMOKE_TEST_IMAGE}:${VERSION}
