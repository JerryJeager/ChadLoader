package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/wader/goutubedl"
)

func DownloadYoutubeVideo(c *gin.Context) {
	url := c.Query("url")
	itagStr := c.Query("itag")
	if url == "" || itagStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url or itag"})
		return
	}

	// Keep itagStr as string for direct comparison with FormatID
	itag, err := strconv.Atoi(itagStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid itag"})
		return
	}

	log.Printf("Fetching video info for: %s (itag: %d)", url, itag)

	// Fetch video info with yt-dlp
	ctx := context.Background()
	vid, err := goutubedl.New(ctx, url, goutubedl.Options{
		Cookies: "/app/cookies.txt",
	})

	if err != nil {
		log.Printf("Error during New: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch video info", "details": err.Error()})
		return
	}
	// Note: No defer vid.Close() - Result doesn't implement io.Closer

	if vid.Info.Title == "" {
		log.Printf("Video title is empty (possible partial fetch failure)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch video metadata"})
		return
	}

	videoTitle := vid.Info.Title
	log.Printf("Got video: %s", videoTitle)

	// Find the format by itag (safely take pointer to slice element)
	var format *goutubedl.Format
	formats := vid.Formats()
	for i := range formats {
		if formats[i].FormatID == itagStr {
			format = &formats[i]
			break
		}
	}

	if format == nil {
		log.Printf("Invalid itag: %d (available IDs: %v)", itag, getFormatIDs(formats))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format or itag"})
		return
	}

	tmpDir := os.TempDir()
	baseName := sanitizeFilename(videoTitle)
	videoPath := filepath.Join(tmpDir, baseName+"_video."+format.Ext)
	audioPath := filepath.Join(tmpDir, baseName+"_audio.m4a")
	outputPath := filepath.Join(tmpDir, baseName+"_final.mp4")

	log.Printf("Downloading format %s (%s %s)", format.FormatID, format.VCodec, format.Resolution)
	// Download video stream
	videoStream, err := vid.Download(ctx, format.FormatID)
	if err != nil {
		log.Printf("Error downloading video stream: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download video", "details": err.Error()})
		return
	}
	defer videoStream.Close()

	// Save to file
	videoFile, err := os.Create(videoPath)
	if err != nil {
		log.Printf("Error creating video file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp file", "details": err.Error()})
		return
	}
	defer videoFile.Close()

	_, err = io.Copy(videoFile, videoStream)
	if err != nil {
		log.Printf("Error saving video: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save video", "details": err.Error()})
		return
	}

	// If no audio, download best audio and merge
	if format.ACodec == "none" {
		log.Println("No audio in this stream, downloading best audio...")

		audioStream, err := vid.Download(ctx, "bestaudio")
		if err != nil {
			log.Printf("Error downloading audio: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download audio", "details": err.Error()})
			return
		}
		defer audioStream.Close()

		audioFile, err := os.Create(audioPath)
		if err != nil {
			log.Printf("Error creating audio file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temp audio", "details": err.Error()})
			return
		}
		defer audioFile.Close()

		_, err = io.Copy(audioFile, audioStream)
		if err != nil {
			log.Printf("Error saving audio: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save audio", "details": err.Error()})
			return
		}

		log.Println("🎬 Merging video and audio with FFmpeg...")
		err = mergeWithFFmpeg(videoPath, audioPath, outputPath)
		if err != nil {
			log.Printf("FFmpeg merge failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to merge video and audio", "details": err.Error()})
			return
		}
	} else {
		log.Println("✅ Stream includes audio, skipping merge.")
		outputPath = videoPath
	}

	// --- Stream file back to client ---
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.mp4\"", baseName))
	c.Header("Content-Type", "video/mp4")

	file, err := os.Open(outputPath)
	if err != nil {
		log.Printf("Failed to open output file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open output file", "details": err.Error()})
		return
	}
	defer file.Close()

	log.Printf("🚀 Streaming final file to client: %s", outputPath)
	_, copyErr := io.Copy(c.Writer, file)
	if copyErr != nil {
		log.Printf("Error streaming file: %v", copyErr)
	}

	// Cleanup
	os.Remove(videoPath)
	if _, err := os.Stat(audioPath); err == nil {
		os.Remove(audioPath)
	}
	if outputPath != videoPath {
		os.Remove(outputPath)
	}

	log.Println("Download completed and temporary files cleaned up.")
}

// Helper to get format IDs for logging
func getFormatIDs(formats []goutubedl.Format) []string {
	ids := make([]string, len(formats))
	for i, f := range formats {
		ids[i] = f.FormatID
	}
	return ids
}

func mergeWithFFmpeg(videoPath, audioPath, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-movflags", "+faststart", // Optimize for streaming
		outputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(out))
	}
	return nil
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalid {
		name = strings.ReplaceAll(name, ch, "_")
	}
	// Truncate if too long (OS limits)
	if len(name) > 100 {
		name = name[:100]
	}
	return name
}
