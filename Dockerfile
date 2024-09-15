# Build the Go binary
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auth-service ./cmd/api

# Image to run the Go binary
FROM alpine:latest
WORKDIR /app
COPY . .
COPY --from=builder /app/auth-service .
RUN apk --no-cache add curl bash
COPY .env .
COPY settings.yml .
COPY wait-for-it.sh .
RUN chmod +x wait-for-it.sh
EXPOSE 1993
ENV PORT 1993
ENV HOSTNAME localhost

CMD ["./wait-for-it.sh", "postgres:5432", "--", "./auth-service"]