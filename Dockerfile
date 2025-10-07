# ---- Build stage ----
FROM golang:1.23-alpine AS builder

# Install required dependencies for Go build
RUN apk add --no-cache ffmpeg python3 py3-pip git

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the app and build
COPY . .
RUN go build -o chadloader main.go


# ---- Final minimal image ----
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ffmpeg python3 py3-pip && \
    pip3 install --break-system-packages --no-cache-dir yt-dlp

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/chadloader .

# Copy cookies.txt 
COPY cookies.txt /app/cookies.txt

# Ensure yt-dlp can find cookies
ENV YT_DLP_COOKIES=/app/cookies.txt

EXPOSE 8080
CMD ["./chadloader"]
