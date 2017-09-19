#!/bin/bash

set -euo pipefail

go get
go build

if ./scopecheck ./testdata > out ; then
    echo "exited cleanly, should have failed"
    exit 1
fi

if ! diff out ./testdata/result.txt > /dev/null ; then
    echo "Output did not match expected result:"
    diff out ./testdata/result.txt
    exit 1
fi
