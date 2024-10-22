#!/bin/bash

set -o pipefail
/app/skupper-ocp-smoke-test -test.v | tee -i /tmp/test.out
ret=$?

/app/go-junit-report < /tmp/test.out | tee /result/junit.xml

exit $ret
