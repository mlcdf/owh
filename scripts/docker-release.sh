#!/bin/bash
# https://docs.docker.com/engine/reference/commandline/build/#custom-build-outputs
    
DOCKER_BUILDKIT=1 docker build -o dist .