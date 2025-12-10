BINARY_NAME=docker-events-exporter
DOCKER_IMAGE=docker-events-exporter

.PHONY: all build clean docker-build run

all: build

# Build for current architecture
build:
	go mod tidy
	go build -o $(BINARY_NAME) main.go

# Build for Linux amd64 (Target)
build-linux:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 main.go

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64

# Run locally for testing (requires local go environment)
run: build
	./$(BINARY_NAME)
