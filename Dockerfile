# Use official golang image as builder
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod tidy
RUN go get github.com/mattn/go-sqlite3

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o syndio-backend

# Use a smaller base image for the final image
FROM alpine:latest

# Install SQLite and required libraries
RUN apk add --no-cache sqlite sqlite-libs

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/syndio-backend .

# Copy the database file
COPY employees.db .

# Expose the default port
EXPOSE 8080

# Run the application
CMD ["./syndio-backend"]
