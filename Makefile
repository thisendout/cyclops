TEST?=./...

default: test

test:
	go test $(TEST) $(TESTARGS)

cover:
	go test $(TEST) -covermode=count -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

build:
	gox -os="darwin linux windows"

build-ci:
	go get github.com/mitchellh/gox
	sudo gox -build-toolchain
	sudo gox -os="darwin linux windows"
