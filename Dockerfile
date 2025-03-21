# Base image
FROM golang:1.24.1 AS builder

# Set working directory
WORKDIR /app

# Copy go mods
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy project source code
COPY . .

# Build Go App for Linux with static linking for alpine
RUN CGO_ENABLED=0 GOOS=linux go build -o /receipt-processor ./cmd

# Create lightweight container image
FROM alpine:latest

# Copy built library over
COPY --from=builder /receipt-processor /receipt-processor

# Expose port
EXPOSE 8080

# Run application
CMD ["/receipt-processor"]