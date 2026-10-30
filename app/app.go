package app

import (
	"Api_Rag/handlers"
	"Api_Rag/routes"
	"Api_Rag/services"
	"Api_Rag/services/gemini"
	"Api_Rag/store"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func NewApp(db *sql.DB, geminiService *gemini.EmbeddingService) *gin.Engine {
	router := gin.Default()

	albumStore := store.NewAlbumStore(db)
	albumService := services.NewAlbumService(albumStore, geminiService)
	albumHandler := handlers.NewAlbumHandler(albumService)

	routes.RegisterRoutes(router, albumHandler)
	return router
}
