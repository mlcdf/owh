#!/bin/bash
set -eu -o pipefail

docker build . \
    -t test \
    -f Dockerfile.test

docker run -it -v $(pwd):/app -e VENOM_VAR_owh=/tests/owh-test test 