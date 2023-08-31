TEST_BINARIES_FOLDER := /image/bin

all: build-tests

build-tests:
	mkdir -p ${TEST_BINARIES_FOLDER}
	go test -c -v ./main/... -o ${TEST_BINARIES_FOLDER}/skupper-ocp-smoke-test
