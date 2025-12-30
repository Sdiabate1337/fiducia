# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install FFmpeg for audio processing
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /build/server .

# Expose port
EXPOSE 8080

# Run
CMD ["./server"]
