package cmd

import (
	"log"
	"os"

	"github.com/JerryJeager/ChadLoader/internal"
	"github.com/JerryJeager/ChadLoader/middleware"
	"github.com/gin-gonic/gin"
)

func ExecuteApiRoutes() {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	api := router.Group("/api/v1/youtube")

	api.GET("", internal.GetYoutubeInfo)
	api.GET("/download", internal.DownloadYoutubeVideo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panic("failed to run server")
	}
}
