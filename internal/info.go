package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kkdai/youtube/v2"
)

func GetYoutubeInfo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing url"})
		return
	}

	client := youtube.Client{}
	video, err := client.GetVideo(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	targetQualities := map[string][]int{
		"360p":  {18, 134, 243},
		"720p":  {22, 136, 247},
		"1080p": {137, 248},
	}

	filtered := []gin.H{}
	addedQualities := map[string]bool{}

	for _, f := range video.Formats {
		for label, itags := range targetQualities {
			if addedQualities[label] {
				continue
			}
			for _, id := range itags {
				if f.ItagNo == id {
					filtered = append(filtered, gin.H{
						"quality": label,
						"itag":    f.ItagNo,
						"type":    f.MimeType,
					})
					addedQualities[label] = true
				}
			}
		}
	}

	var bestAudio *youtube.Format
	for _, f := range video.Formats {
		if f.AudioChannels > 0 && f.QualityLabel == "" { 
			if bestAudio == nil || f.Bitrate > bestAudio.Bitrate {
				bestAudio = &f
			}
		}
	}

	if bestAudio != nil {
		filtered = append(filtered, gin.H{
			"quality": "audio",
			"itag":    bestAudio.ItagNo,
			"type":    bestAudio.MimeType,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"title":       video.Title,
		"author":      video.Author,
		"description": video.Description,
		"thumbnails":  video.Thumbnails,
		"date":        video.PublishDate,
		"formats":     filtered,
	})
}
