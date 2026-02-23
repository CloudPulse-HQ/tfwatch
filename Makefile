.PHONY: build install test clean run list docker-up docker-down publish-examples

BINARY_NAME=tfwatch
VERSION=1.0.0
BUILD_DIR=build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

install: build
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed"

test:
	go test -v ./...

run: build
	./$(BUILD_DIR)/$(BINARY_NAME) --dir ./examples/eks-cluster

list: build
	./$(BUILD_DIR)/$(BINARY_NAME) --list --dir ./examples/eks-cluster

docker-up:
	docker compose -f deploy/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

publish-examples: build
	@for dir in examples/*/; do \
		echo "Publishing $$dir..."; \
		./$(BUILD_DIR)/$(BINARY_NAME) --dir "$$dir" || true; \
	done

clean:
	rm -rf $(BUILD_DIR)
