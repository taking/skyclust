# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY internal/ ./internal/
COPY cmd/ ./cmd/

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cmp-server cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/cmp-server .

# Copy plugins directory
COPY plugins/ ./plugins/

# Copy config file
COPY config.yaml .

# Create non-root user
RUN adduser -D -s /bin/sh cmp
RUN chown -R cmp:cmp /app
USER cmp

EXPOSE 8080

CMD ["./cmp-server", "run"]
