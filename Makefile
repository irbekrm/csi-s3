UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)
	OS := "darwin"
else ifeq($(UNAME_S),linux)
	OS := "linux"
else
$(error Unsupported OS: $(UNAME_S))
endif

.PHONY: build test clean

build:
	GOOS=linux GOARCH=amd64 go build -o outputs/csi-s3

test:
	./hack/test_$(OS).sh

clean:
	rm -f outputs/*