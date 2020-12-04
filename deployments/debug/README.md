## debug

Debug/development deployment of csi-s3 driver

Deploys a pod with two containers:

- A container running the latest version of `csi-s3`
- A container with (`grpcurl`)[https://github.com/fullstorydev/grpcurl]

The Unix Domain Socket via which the gRPC API of `csi-s3` can be accessed is also mounted in the container with `grpcurl` at `/csi`. 


### Usage

#### Deploy

`kubectl apply -f deployments/debug/deployment.yaml`

#### Exec into the 'grpcurl' container

`kubectl exec -it <POD> -c grpcurl -- sh`

#### Execute `grpcurl` commands to against `csi-s3` running into the other container

i.e `./grpcurl -unix -plaintext /csi/csi.sock list`

#### Example bucket mount

- `kubectl exec -it <POD> -c grpcurl -- sh`

- Mount by running
```
./grpcurl -unix \
  -plaintext \
  -d='{"volumeId":"<YOUR-BUCKET-NAME>","secrets":{"AWS_ACCESS_KEY_ID":"<AWS-ACCESS-KEY-ID>","AWS_SECRET_ACCESS_KEY":"<AWS-SECRET-ACCESS-KEY>"},"targetPath":"<TARGET-PATH>"}' \
  /csi/csi.sock csi.v1.Node.NodePublishVolume
```

- Unmount by running
```
  ./grpcurl -unix \
  -plaintext \
  -d='{"volumeId":"irbe-csi-test","targetPath":"<TARGET-PATH>"}' \
  /csi/csi.sock csi.v1.Node.NodeUnpublishVolume
```