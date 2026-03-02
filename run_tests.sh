#!/bin/bash

set -e

TEST_DIR="test"

if [ ! -d "$TEST_DIR" ]; then
    echo "Error: test directory not found"
    exit 1
fi

echo "Running all tests in $TEST_DIR..."
go test ./test/... -v

echo "All tests completed successfully!"