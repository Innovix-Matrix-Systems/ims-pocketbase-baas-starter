# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy public files
COPY --from=builder /app/pb_public ./pb_public

# Copy schema files for migrations
COPY --from=builder /app/internal ./internal

# Create pb_data directory
RUN mkdir -p pb_data

# Expose port
EXPOSE 8090

# Run the binary
CMD ["./main", "serve", "--http=0.0.0.0:8090"]