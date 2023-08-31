FROM golang:1.19 AS builder

WORKDIR /image
COPY src/go.mod .
COPY src/go.sum .
RUN go mod download
COPY . .

WORKDIR /image/src
RUN make all
RUN go get github.com/jstemmer/go-junit-report
RUN go install github.com/jstemmer/go-junit-report

FROM registry.access.redhat.com/ubi8-minimal

RUN mkdir -p /config /result
RUN chgrp -R 0 /result && chmod -R g=u /result
WORKDIR /app

COPY --from=builder /image/bin/skupper-ocp-smoke-test .
COPY --from=builder /go/bin/go-junit-report .
COPY run-test.sh .
CMD './run-test.sh'

