# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build server and CLI
RUN go build -o server cmd/server/main.go
RUN go build -o deployment-tail cmd/cli/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/deployment-tail .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/web ./web

# Expose API port
EXPOSE 8080

# Run server by default
CMD ["./server"]
