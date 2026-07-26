# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy dependency files first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the main API server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api-server ./cmd/main.go

# Build the judge worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/judge-worker ./cmd/judge/main.go

# Build the database seeder
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seeder ./cmd/seed/main.go

# Runtime stage
FROM alpine:3.20

WORKDIR /app

# Install ca-certificates and tzdata
RUN apk add --no-cache ca-certificates tzdata

# Copy built binaries from builder
COPY --from=builder /app/api-server /app/api-server
COPY --from=builder /app/judge-worker /app/judge-worker
COPY --from=builder /app/seeder /app/seeder

# Copy .env.example as a fallback .env if not mounted
COPY .env.example /app/.env

# Default command starts the API server
CMD ["/app/api-server"]
