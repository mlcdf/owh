#!/usr/bin/env bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd .. && go test . -tags testbincover -coverpkg="./..." -c -o "${SCRIPT_DIR}/../dist/owh.test" -ldflags "-X go.mlcdf.fr/owh/internal/test.IsTest=true"