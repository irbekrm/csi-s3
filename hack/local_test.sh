#!/bin/bash

set -eux

# build an image with the contents of the repo
docker build -t "csi-s3/tests:latest" -f build/Dockerfile.test .

# run the image just built, streaming output to the terminal
docker run -t "csi-s3/tests:latest"