# csi-s3
CSI Driver for S3 via FUSE

## Development
### Tests

This project can only be built for Linux OS because of a dependency on a Linux-specific filesystem package.

To run the unit tests (using Docker) on any OS:

1. Ensure [Docker](https://docs.docker.com/get-docker/) is installed and running

2. From the root of repository run `./scripts/local_test.sh`

If you have made any code changes, you might also want to regenerate the mocks (see below)

### Build

It is only possible to build for Linux targets.

Run `GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3`

### Mocks

This project uses generated [gomock](https://github.com/golang/mock) mocks for unit testing. The generated mocks are in the /mocks repo

To regenerate the mocks:

1. Install mockgen `go get github.com/golang/mock/mockgen@v1.4.4`

2. Run `GOOS=linux GOARCH=amd64 go generate ./...`