FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /pb main.go

# Runtime stage
FROM alpine:latest

WORKDIR /

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /pb /pb

# Expose port
EXPOSE 8090

# Define volume for data
VOLUME /pb_data

# Start PocketBase
# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start PocketBase
ENTRYPOINT ["/entrypoint.sh"]
