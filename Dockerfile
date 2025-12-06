# Stage 1: Builder
FROM golang:alpine AS builder

# Install system dependencies required for CGO or specific tools if needed
# fiber/imaging is pure Go largely, but 'imaging' might need minimal dependencies
RUN apk add --no-cache git

WORKDIR /app

# Download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 creates a statically linked binary (no dependency on system libc)
RUN CGO_ENABLED=0 GOOS=linux go build -o cdn-service main.go

# Stage 2: Runner
FROM alpine:latest

# Install CA certificates for SSL and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/cdn-service .

# Create data directories structure
RUN mkdir -p data/avatar data/image data/cache

# Expose the application port
EXPOSE 3000

# Run the binary
CMD ["./cdn-service"]
