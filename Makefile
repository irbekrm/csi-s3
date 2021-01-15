
.PHONY: build
build:
	GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3

.PHONY: test
test:
	./hack/local_test.sh

.PHONY: clean
clean:
	rm -f outputs/*