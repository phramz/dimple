.PHONY: vendors
vendors:
	go mod download

.PHONY: format
format:
	go fmt ./...

.PHONY: lint
lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint golangci-lint run

.PHONY: test
test:
	go test -race -cover ./...
