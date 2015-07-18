TEST?=./...

default: test

test:
	go test $(TEST) $(TESTARGS)

cover:
	go test $(TEST) -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

build:
	go get github.com/mitchellh/gox
	gox -build-toolchain
	gox -os="darwin linux windows"
