#!/bin/bash
set -eu -o pipefail

docker run --name owh-test --rm -d -i -t -v $(pwd):/app golang:1.19.3-bullseye bash
docker exec -w /app owh-test go test ./...
docker kill owh-test