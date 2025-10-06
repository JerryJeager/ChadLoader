FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ffmpeg python3 py3-pip

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o chadloader main.go


# --- Final minimal image ---
FROM alpine:latest

RUN apk add --no-cache ffmpeg python3 py3-pip && \
    pip3 install --break-system-packages --no-cache-dir yt-dlp

WORKDIR /app

# Copy the built binary
COPY --from=builder /app/chadloader .

# Copy cookies.txt for yt-dlp authentication
COPY cookies.txt /app/cookies.txt

EXPOSE 8080
CMD ["./chadloader"]
