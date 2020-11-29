## debug

Debug/development deployment of csi-s3 driver

Deployment with a pod with two containers:

- A container running latest version of `csi-s3`
- A container running (`grpcurl`)[https://github.com/fullstorydev/grpcurl]

The Unix Domain Socket via which the gRPC API of `csi-s3` can be accessed is also mounted in the container with `grpcurl` at `/csi`


### Usage

#### Deploy

`kubectl apply -f deployments/debug/deployment.yaml`

#### Exec into the container with `grpcurl`

`kubectl exec -it <POD> -c grpcurl -- sh`

#### Execute `grpcurl` commands to against `csi-s3` running into the other container

i.e `./grpcurl -unix -plaintext /csi/csi.sock list`