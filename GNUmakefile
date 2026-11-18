default: lint install generate

build:
	go build -v .

install: build
	go install -v .

lint:
	golangci-lint run --fix

generate:
	go generate ./...

TEST_ARGS ?= ./...

test:
	go test -v -cover -timeout=30s -parallel=8 $(TEST_ARGS)

testacc:
	TF_ACC=1 go test -v -cover -timeout=30m -parallel=8 $(TEST_ARGS)

.PHONY: build install lint generate test testacc
