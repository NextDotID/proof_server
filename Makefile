.PHONY: build test

bin_dir=build/

build:
	@go build -o ${bin_dir} ./cmd/...

test:
	@go test -v ./...
