#!/bin/bash

set -e

# note: tests will run with 'latest' version of xdeb for now
#   because tests/test_xdeb.py installs 'latest' last

mkdir -p html
PYTEST_TEST_REPORT_ARGS="--html=html/test-report.html --self-contained-html"

if [ "${1}" = "html" ]; then
    pytest ${PYTEST_TEST_REPORT_ARGS}
else
    pytest
fi
