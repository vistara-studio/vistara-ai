# Build stage
FROM golang:1.24.5-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for private repos and SSL)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/api/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for SSL
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy .env file if exists (optional)
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 5000

# Run the application
CMD ["./main"]
