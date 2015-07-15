TEST?=./...

default: test

test:
	go test $(TEST) $(TESTARGS)

cover:
	go test $(TEST) -coverprofile=coverage.out
	go tool cover -html=coverage.out
	rm coverage.out

