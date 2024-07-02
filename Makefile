.PHONY: generate
generate:
	go generate ./...

.PHONY: format
format:
	golangci-lint run --fix ./...

.PHONY: check
check:
	golangci-lint run ./...

.PHONY: test
test:
	go test -v ./...
