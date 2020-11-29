FROM golang:alpine AS build

WORKDIR /app

COPY . .

RUN ls
RUN go build -o outputs/csi-s3

# From https://github.com/efrecon/docker-s3fs-client/blob/master/Dockerfile
ARG S3FS_VERSION=v1.86
RUN apk --no-cache add \
    ca-certificates \
    build-base \
    git \
    alpine-sdk \
    libcurl \
    automake \
    autoconf \
    libxml2-dev \
    libressl-dev \
    fuse-dev \
    curl-dev && \
  git clone https://github.com/s3fs-fuse/s3fs-fuse.git && \
  cd s3fs-fuse && \
  git checkout tags/${S3FS_VERSION} && \
  ./autogen.sh && \
  ./configure --prefix=/usr && \
  make -j && \
  make install

FROM alpine

COPY --from=build /usr/bin/s3fs /usr/bin/s3fs
COPY --from=build /app/outputs/csi-s3 /usr/bin/csi-s3

ENTRYPOINT ["csi-s3", "--csi-address=/csi/csi.sock"]