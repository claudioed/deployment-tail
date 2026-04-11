.PHONY: build test test-integration test-bdd test-bdd-smoke test-bdd-service test-bdd-http test-bdd-ui test-bdd-ci run-server run-cli migrate clean docker-up docker-down mutation-install mutation-test mutation-test-dry

# Build both server and CLI
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/deployment-tail cmd/cli/main.go

# Run tests (excluding BDD tests - use test-bdd for those)
test:
	go test -v $(shell go list ./... | grep -v '/test/bdd')

# Run integration tests
test-integration:
	go test -v -tags=integration ./...

# BDD tests (Behavior-Driven Development with Godog)
test-bdd:
	go test -v -count=1 ./test/bdd -- -godog.format=pretty -godog.tags="~@wip"

test-bdd-smoke:
	go test -v -count=1 ./test/bdd -- -godog.tags="@smoke && ~@wip" -godog.format=pretty

test-bdd-service:
	go test -v -count=1 ./test/bdd -- -godog.tags="@service && ~@wip"

test-bdd-http:
	go test -v -count=1 ./test/bdd -- -godog.tags="@http && ~@wip"

test-bdd-ui:
	go test -v -count=1 ./test/bdd -- -godog.tags="@ui && ~@wip"

test-bdd-ci:
	go test -v -count=1 ./test/bdd -- -godog.format="progress,junit:bdd-report.xml" -godog.tags="~@wip"

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

# Mutation testing (Gremlins) - https://gremlins.dev
GREMLINS_VERSION ?= 0.6.0

# Install Gremlins if not already on $PATH
mutation-install:
	@if command -v gremlins >/dev/null 2>&1; then \
		echo "gremlins already installed: $$(gremlins --version)"; \
	else \
		echo "gremlins not found on \$$PATH."; \
		echo "Install v$(GREMLINS_VERSION) from https://github.com/go-gremlins/gremlins/releases"; \
		echo "  macOS (Apple Silicon):"; \
		echo "    curl -sSL https://github.com/go-gremlins/gremlins/releases/download/v$(GREMLINS_VERSION)/gremlins_$(GREMLINS_VERSION)_darwin_arm64.tar.gz | tar -xz -C /tmp gremlins && sudo install -m 0755 /tmp/gremlins /usr/local/bin/gremlins"; \
		echo "  Linux (amd64):"; \
		echo "    curl -sSL https://github.com/go-gremlins/gremlins/releases/download/v$(GREMLINS_VERSION)/gremlins_$(GREMLINS_VERSION)_linux_amd64.tar.gz | tar -xz -C /tmp gremlins && sudo install -m 0755 /tmp/gremlins /usr/local/bin/gremlins"; \
		exit 1; \
	fi

# Run the full mutation testing campaign (fails if thresholds are not met)
mutation-test: mutation-install
	@echo "Running mutation tests on all internal packages..."
	@for pkg in $$(go list ./internal/...); do \
		echo "==> Testing $$pkg"; \
		gremlins unleash "./$${pkg#github.com/claudioed/deployment-tail/}" || exit 1; \
	done

# Dry-run: list mutants without running tests (fast iteration)
mutation-test-dry: mutation-install
	@echo "Listing mutants for all internal packages..."
	@for pkg in $$(go list ./internal/...); do \
		echo "==> $$pkg"; \
		gremlins unleash --dry-run "./$${pkg#github.com/claudioed/deployment-tail/}"; \
	done

# Domain-only mutation testing (fast, high-coverage core business logic)
mutation-test-domain: mutation-install
	@echo "Running mutation tests on domain packages..."
	@for pkg in $$(go list ./internal/domain/...); do \
		echo "==> Testing $$pkg"; \
		gremlins unleash "./$${pkg#github.com/claudioed/deployment-tail/}" || exit 1; \
	done
