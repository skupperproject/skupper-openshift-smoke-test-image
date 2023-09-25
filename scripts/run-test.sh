#!/bin/bash
/app/skupper-ocp-smoke-test -test.v | tee /tmp/test.out
cat /tmp/test.out | /app/go-junit-report | tee /result/junit.xml
