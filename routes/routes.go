package routes

import (
	"Api_Rag/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, albumHandler *handlers.AlbumHandler) {
	router.GET("/albums", albumHandler.GetAlbums)
	router.GET("/albums/:id", albumHandler.GetAlbumByID)
	router.POST("/albums", albumHandler.PostAlbums)
	router.POST("/albums/search", albumHandler.SearchSimilar)
}
