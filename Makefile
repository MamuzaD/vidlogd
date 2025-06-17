BINARY_NAME=vidlogd

# go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
MAKEFLAGS += --silent

.PHONY: all build clean run

# build:
all: build

build:
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(GOBIN)
	go build -o $(GOBIN)/$(BINARY_NAME)
	@echo "Built $(BINARY_NAME) in $(GOBIN)"

clean:
	@echo "Cleaning..."
	rm -rf $(GOBIN)
	go clean
	@echo "Cleaned $(BINARY_NAME) in $(GOBIN)"

run:
	go run .
