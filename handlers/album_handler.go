package handlers

import (
	"Api_Rag/models"
	"Api_Rag/services"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlbumHandler struct {
	service *services.AlbumService
}

func NewAlbumHandler(service *services.AlbumService) *AlbumHandler {
	return &AlbumHandler{
		service: service,
	}
}

func (h *AlbumHandler) GetAlbums(c *gin.Context) {
	albums, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, albums)
}

func (h *AlbumHandler) GetAlbumByID(c *gin.Context) {
	id := c.Param("id")
	album, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, album)
}

func (h *AlbumHandler) PostAlbums(c *gin.Context) {
	var album models.Album
	if err := c.ShouldBindJSON(&album); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.Create(c.Request.Context(), album)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

type SearchRequest struct {
	Query         string  `json:"query" binding:"required"`
	Limit         int     `json:"limit"`
	MinSimilarity float64 `json:"min_similarity"` // Ex: 0.60 ou 0.70
}

// SearchSimilar recebe uma pergunta em texto natural, gera o embedding via Gemini e busca no pgvector
func (h *AlbumHandler) SearchSimilar(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.service.SearchSimilar(c.Request.Context(), req.Query, req.Limit, req.MinSimilarity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}
