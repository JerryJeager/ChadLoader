package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wader/goutubedl"
)

func GetYoutubeInfo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := goutubedl.New(ctx, url, goutubedl.Options{
		Cookies: "/app/cookies.txt",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	info := result.Info

	targetQualities := map[string][]int{
		"360p":  {18, 134, 243},
		"720p":  {22, 136, 247},
		"1080p": {137, 248},
	}

	filtered := []gin.H{}
	addedQualities := map[string]bool{}

	for _, f := range info.Formats {
		formatID := 0
		if f.FormatID != "" {
			// Parse format ID string to int
			var id int
			if _, err := fmt.Sscanf(f.FormatID, "%d", &id); err == nil {
				formatID = id
			}
		}

		for label, itags := range targetQualities {
			if addedQualities[label] {
				continue
			}
			for _, id := range itags {
				if formatID == id {
					filtered = append(filtered, gin.H{
						"quality": label,
						"itag":    formatID,
						"type":    f.Ext,
					})
					addedQualities[label] = true
				}
			}
		}
	}

	// Find best audio format
	var bestAudio *goutubedl.Format
	for i, f := range info.Formats {
		// Audio-only formats typically have VCodec as "none"
		if f.VCodec == "none" && f.ACodec != "none" {
			if bestAudio == nil || f.ABR > bestAudio.ABR {
				bestAudio = &info.Formats[i]
			}
		}
	}

	if bestAudio != nil {
		formatID := 0
		if bestAudio.FormatID != "" {
			fmt.Sscanf(bestAudio.FormatID, "%d", &formatID)
		}
		filtered = append(filtered, gin.H{
			"quality": "audio",
			"itag":    formatID,
			"type":    bestAudio.Ext,
		})
	}

	// Build thumbnails array
	thumbnails := []gin.H{}
	for _, thumb := range info.Thumbnails {
		thumbnails = append(thumbnails, gin.H{
			"url":    thumb.URL,
			"width":  thumb.Width,
			"height": thumb.Height,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       info.Title,
		"author":      info.Uploader,
		"description": info.Description,
		"thumbnails":  thumbnails,
		"date":        info.UploadDate,
		"formats":     filtered,
	})
}
