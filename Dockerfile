FROM golang:1.19 AS builder

WORKDIR /image
COPY . .

RUN go mod download
RUN make build-tests
RUN go get github.com/jstemmer/go-junit-report
RUN go install github.com/jstemmer/go-junit-report

FROM registry.access.redhat.com/ubi8-minimal

RUN mkdir -p /config /result
RUN chgrp -R 0 /result && chmod -R g=u /result
WORKDIR /app

COPY --from=builder /image/bin/skupper-ocp-smoke-test .
COPY --from=builder /go/bin/go-junit-report .
COPY scripts/run-test.sh .
CMD './run-test.sh'

