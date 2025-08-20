BINARY_NAME=vidlogd

# go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
MAKEFLAGS += --silent

.PHONY: all build clean run release

# build:
all: build

build:
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(GOBIN)
	go build -o $(GOBIN)/$(BINARY_NAME) ./cmd/vidlogd
	@echo "Built $(BINARY_NAME) in $(GOBIN)"

clean:
	@echo "Cleaning..."
	rm -rf $(GOBIN)
	go clean
	@echo "Cleaned $(BINARY_NAME) in $(GOBIN)"

run:
	go run ./cmd/vidlogd

release: clean
	@echo "Building releases for multiple platforms..."
	# macOS
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 ./cmd/vidlogd
	# Linux
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/$(BINARY_NAME)-linux-amd64 ./cmd/vidlogd
	# Windows
	GOOS=windows GOARCH=amd64 go build -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe ./cmd/vidlogd
	@echo "Built releases in $(GOBIN)"
