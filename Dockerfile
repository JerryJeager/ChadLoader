FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ffmpeg python3 py3-pip

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o chadloader main.go

FROM alpine:latest

RUN apk add --no-cache ffmpeg python3 py3-pip && \
    pip3 install --no-cache-dir yt-dlp

WORKDIR /app

COPY --from=builder /app/chadloader .

EXPOSE 8080

CMD ["./chadloader"]