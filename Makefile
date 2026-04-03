.PHONY: all build test lint vet fmt clean cover install

# Default
all: fmt lint vet test build

# Build binary
build:
	cd hyoka && go build -o ../bin/hyoka .

# Run tests with race detection
test:
	cd hyoka && go test -race -count=1 ./...

# Lint with golangci-lint
lint:
	cd hyoka && golangci-lint run ./...

# Go vet
vet:
	cd hyoka && go vet ./...

# Format and tidy
fmt:
	cd hyoka && gofmt -w . && go mod tidy

# Coverage report
cover:
	cd hyoka && go test -race -coverprofile=coverage.txt -covermode=atomic ./... && go tool cover -func=coverage.txt

# Install binary
install: build
	cp bin/hyoka $(GOPATH)/bin/hyoka 2>/dev/null || cp bin/hyoka /usr/local/bin/hyoka

# Clean
clean:
	rm -rf bin/ hyoka/coverage.txt
