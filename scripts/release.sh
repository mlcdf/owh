#!/bin/bash
set -eu -o pipefail
    
VERSION=$(git rev-parse --short HEAD)
OS="linux"
ARCH="amd64"

OUTPUT="./dist/owh-${VERSION}-${OS}-${ARCH}"

# https://golang.org/cmd/link/
go build -o $OUTPUT -tags release -ldflags="-X 'main.Version=${VERSION}'"
gzip --force --keep -v $OUTPUT