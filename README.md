# csi-s3
CSI Driver for S3 via FUSE

:warning: This project is currently in pre-alpha state and is not yet usable. See [below](#deploying-on-kubernetes) for how to try it out :warning:

## Description

A Kubernetes [CSI driver](https://kubernetes.io/blog/2019/01/15/container-storage-interface-ga/) for creating [Persistent Volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) backed by S3.

## Supported mounters

- [s3fs](https://github.com/s3fs-fuse/s3fs-fuse)

## Supported S3 types
- AWS S3 (at the moment only pre-existing buckets)
## Implementation
### Kubernetes

`csi-s3` is implemented according to the [CSI spec](https://github.com/container-storage-interface/spec/blob/master/spec.md).

It exposes a gRPC API over a Unix Domain Socket. The RPCs in this API are called by the kubelet as well as the various [CSI sidecar containers](https://kubernetes-csi.github.io/docs/sidecar-containers.html).

`csi-s3` has to be deployed as a Daemonset (it needs to be running on the node to be able to mount the volume)

### Mounting

Mounting S3 to filesystem is possible via [FUSE](https://en.wikipedia.org/wiki/Filesystem_in_Userspace).

`csi-s3` invokes [higher level tools](#supported-mounters) that do the actual mounting.
## Development
### Tests

This project can only be built for Linux targets because of a dependency on a Linux-specific filesystem package. 

To run the unit tests (using Docker) on any OS:

1. Ensure [Docker](https://docs.docker.com/get-docker/) is installed and running

2. From the root of repository run `./scripts/local_test.sh`

If you have made any code changes, you might also want to regenerate the [mocks](#mocks)

### Build

It is only possible to build for Linux targets.

Run `GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3`

### Mocks

This project uses generated [gomock](https://github.com/golang/mock) mocks for unit testing. The generated mocks are at `/mocks`

To regenerate the mocks:

1. Install mockgen `go get github.com/golang/mock/mockgen@v1.4.4`

2. Run `GOOS=linux GOARCH=amd64 go generate ./...`

### Deploying on Kubernetes

1. Deploy `csi-s3` driver (as a Daemonset), RBAC resources and a `CSIDriver` custom resource

`kubectl apply -f deployments/`

2. See [/examples](examples/README.md) for how to create a Persistent Volume backed by `csi-s3` and use it

### Manually testing the API

See [/deployments/debug](deployments/debug/README.md) for an example of how to run `csi-s3` and manually test the API.

### CSI Compatibility

The gRPC API of `csi-s3` implements a subset of the functionality described by the [CSI spec](https://github.com/container-storage-interface/spec/blob/master/spec.md)

**Currently implemented RPCs from the CSI spec are:**

- Node Service
   - [NodePublishVolume](https://github.com/container-storage-interface/spec/blob/master/spec.md#nodepublishvolume) RPC - mounts an already existing bucket
   - [NodeUnpublishVolume](https://github.com/container-storage-interface/spec/blob/master/spec.md#nodeunpublishvolume) RPC - unmounts a bucket
   - [NodeGetInfo](https://github.com/container-storage-interface/spec/blob/master/spec.md#nodegetinfo) RPC - node id (from plugin's perspective)
   - [NodeGetCapabilities](https://github.com/container-storage-interface/spec/blob/master/spec.md#nodegetcapabilities) RPC- optional node capabilities that the driver implements

- Identity Service

    - [Probe](https://github.com/container-storage-interface/spec/blob/master/spec.md#probe) RPC - verifies that the driver is healthy
    - [GetPluginInfo](https://github.com/container-storage-interface/spec/blob/master/spec.md#getplugininfo) RPC - returns name and version of the driver
    - [GetPluginCapabilities](https://github.com/container-storage-interface/spec/blob/master/spec.md#getplugincapabilities) RPC - additional capabilities/constraints of the driver