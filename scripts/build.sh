#!/bin/bash

set -eux

tag=${TAG:-test}

docker build -t "irbekrm/csi-s3:${tag}" -t irbekrm/csi-s3:latest -f build/Dockerfile .

docker push irbekrm/csi-s3