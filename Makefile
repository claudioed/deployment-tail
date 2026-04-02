.PHONY: build test run-server run-cli migrate clean docker-up docker-down

# Build both server and CLI
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/deployment-tail cmd/cli/main.go

# Run tests
test:
	go test -v ./...

# Run integration tests
test-integration:
	go test -v -tags=integration ./...

# Run the API server
run-server:
	go run cmd/server/main.go

# Run database migrations
migrate:
	go run cmd/server/main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Start Docker Compose (MySQL)
docker-up:
	docker-compose up -d

# Stop Docker Compose
docker-down:
	docker-compose down

# Install CLI locally
install:
	go install cmd/cli/main.go

# Generate OpenAPI stubs
generate:
	go generate ./api/...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Development setup
dev-setup: docker-up
	@echo "Waiting for MySQL to be ready..."
	@sleep 5
	@echo "Running migrations..."
	@DB_HOST=localhost DB_PORT=3306 DB_USER=root DB_PASSWORD=rootpass DB_NAME=deployment_schedules go run cmd/server/main.go &
	@echo "Development environment ready!"
