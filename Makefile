.PHONY: build test clean install lint fmt run

BINARY_NAME=graphql-go-gen
BINARY_PATH=./cmd/graphql-go-gen

build:
	go build -o ${BINARY_NAME} ${BINARY_PATH}

install:
	go install ${BINARY_PATH}

test:
	go test -v -race -cover ./...

test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	go clean
	rm -f ${BINARY_NAME}
	rm -f coverage.out coverage.html

lint:
	@echo "Running linters..."
	go vet ./...
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt:
	go fmt ./...
	goimports -w .

deps:
	go mod download
	go mod tidy

run: build
	./${BINARY_NAME}

bench:
	go test -bench=. -benchmem ./...

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  install      - Install the binary to GOPATH/bin"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  run          - Build and run"
	@echo "  bench        - Run benchmarks"