# csi-s3
CSI Driver for S3 via FUSE

## Development
### Tests

This project can only be built for Linux OS because of a dependency on a Linux-specific filesystem package. (It is intended to be used only on Linux OS)

To run the unit tests (using Docker) on any OS:

1. Ensure [Docker](https://docs.docker.com/get-docker/) is installed and running

2. From the root of repository run `./scripts/local_test.sh`