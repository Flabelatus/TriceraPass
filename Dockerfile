# Stage 1: Build the Go binary
FROM golang:1.19-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application (static binary for Linux)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-service ./cmd/api

# Stage 2: Run the Go binary in a minimal image
FROM alpine:latest

# Install curl and bash in the Alpine image (bash is required for wait-for-it.sh)
RUN apk --no-cache add curl bash

# Set the working directory for the runtime environment to /app
WORKDIR /app

# Copy the rest of the application code
COPY . .

# Copy the binary from the builder stage to /app
COPY --from=builder /app/auth-service .

# Copy .env and settings.yml to /app
COPY .env .
COPY settings.yml .

# Copy the wait-for-it.sh script to /app and make it executable
COPY wait-for-it.sh .
RUN chmod +x wait-for-it.sh

# Expose the application's port (optional)
EXPOSE 1993
ENV PORT 1993
ENV HOSTNAME localhost

# Run the wait-for-it.sh script to ensure Postgres is ready, then start the Go binary
CMD ["./wait-for-it.sh", "postgres:5432", "--", "./auth-service"]