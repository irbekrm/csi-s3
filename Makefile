.PHONY: build test clean
build:
	GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3

test:
	./hack/local_test.sh

clean:
	rm -f outputs/*