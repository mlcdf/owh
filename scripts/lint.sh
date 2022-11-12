#!/bin/bash
set -eu -o pipefail

go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

# decomment for debug
# export GL_DEBUG=loader,gocritic 
golangci-lint run ./... -v