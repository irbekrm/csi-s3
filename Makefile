UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)
	OS := "darwin"
else ifeq ($(UNAME_S),Linux)
	OS := "linux"
else
$(error Unsupported OS: $(UNAME_S))
endif

.PHONY: all build test clean update fmt generate vet tidy

all: update test build

build:
	GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3

test: vet
	./hack/test_$(OS).sh

update: tidy fmt generate

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

generate:
	GOOS=linux GOARCH=amd64 go generate ./...

clean:
	rm -f outputs/*