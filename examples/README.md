# example usage of csi-s3

Create a secret with AWS creds

```
kubectl create secret generic csi-s3 \
    --from-literal=AWS_ACCESS_KEY_ID=<AWS-ACCESS-KEY-ID> \
    --from-literal=AWS_SECRET_ACCESS_KEY=<AWS-SECRET-ACCESS-KEY>
```

Create a Persistent Volume representing the bucket you want to mount

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolume
metadata:
  name: csi-s3-pv
spec:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 1Gi
  storageClassName: s3.csi.irbe.dev
  csi:
    driver: s3.csi.irbe.dev
    volumeHandle: <BUCKET-NAME>
    nodePublishSecretRef:
      name: csi-s3
      namespace: default
EOF
```

Create a Persistent Volume Claim for the volume above. Check that it gets bound to the PV

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: csi-s3-pvc
spec:
  volumeName: csi-s3-pv
  accessModes:
  - ReadWriteOnce
  storageClassName: s3.csi.irbe.dev
  resources:
    requests:
      storage: 1G
EOF
```

Create an example deployment that uses the PVC. 

```
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: csi-s3-test
spec:
  selector:
    matchLabels:
      app: csi-s3-test
  template:
    metadata:
      labels:
        app: csi-s3-test
    spec:
      containers:
      - name: csi-s3-test
        image: busybox
        command:
        - sleep
        - infinity
        volumeMounts:
        - name: csi-s3-pvc
          mountPath: /data
      volumes:
      - name: csi-s3-pvc
        persistentVolumeClaim:
          claimName: csi-s3-pvc
EOF
```

You should now have RW access to the bucket via `/data` directory in the `csi-s3-test` bucket