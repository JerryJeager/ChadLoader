# ChadLoader

ChadLoader is a simple Go-based REST API server for fetching YouTube video information and downloading videos. It uses the Gin web framework and supports CORS for easy integration with web clients.

## Features

- Fetch YouTube video metadata and available formats
- Download YouTube videos in selected formats
- Automatically merges video and audio streams if needed (using FFmpeg)
- CORS support for cross-origin requests

## API Endpoints

### 1. Health Check

**GET /**  
Returns a simple JSON message.

**Response:**
```json
{ "message": "Hello World" }
```

### 2. Get YouTube Video Info

**GET /api/v1/youtube?url=VIDEO_URL**

Fetches metadata and available formats for a YouTube video.

**Query Parameters:**
- `url` (required): The YouTube video URL

**Response:**
```json
{
  "title": "...",
  "author": "...",
  "description": "...",
  "thumbnails": [...],
  "date": "...",
  "formats": [
    { "quality": "720p", "itag": 22, "type": "video/mp4; codecs=\"...\"" },
    { "quality": "audio", "itag": 140, "type": "audio/mp4; codecs=\"...\"" }
    // ...
  ]
}
```

### 3. Download YouTube Video

**GET /api/v1/youtube/download?url=VIDEO_URL&itag=ITAG**

Downloads the selected format. If the format lacks audio, the best audio stream is merged using FFmpeg.

**Query Parameters:**
- `url` (required): The YouTube video URL
- `itag` (required): The format itag (from `/api/v1/youtube` response)

**Response:**
- Streams the final video file (`.mp4`) as an attachment.

## Setup

### Prerequisites

- Go 1.23.3+
- [FFmpeg](https://ffmpeg.org/) installed and available in your system PATH

### Install Dependencies

```sh
go mod tidy
```

### Run the Server

```sh
go run main.go
```

The server listens on port `8080` by default (or set `PORT` environment variable).

## Project Structure

- [`main.go`](main.go): Entry point, starts the server
- [`cmd/app.go`](cmd/app.go): API route setup
- [`internal/info.go`](internal/info.go): YouTube info endpoint
- [`internal/download.go`](internal/download.go): Download and merge logic
- [`middleware/cors.go`](middleware/cors.go): CORS middleware

## License

MIT

---

**Note:** This project uses [`github.com/kkdai/youtube/v2`](https://github.com/kkdai/youtube) for metadata and [`github.com/wader/goutubedl`](https://github.com/wader/goutubedl) for downloads.