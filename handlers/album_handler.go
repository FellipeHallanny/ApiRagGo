package handlers

import (
	"Api_Rag/models"
	"Api_Rag/services/gemini"
	"Api_Rag/store"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AlbumHandler struct {
	store         *store.AlbumStore
	geminiService *gemini.EmbeddingService
}

func NewAlbumHandler(store *store.AlbumStore, geminiService *gemini.EmbeddingService) *AlbumHandler {
	return &AlbumHandler{
		store:         store,
		geminiService: geminiService,
	}
}

func (h *AlbumHandler) GetAlbums(c *gin.Context) {
	albums, err := h.store.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if albums == nil {
		albums = []models.Album{}
	}
	c.JSON(http.StatusOK, albums)
}

func (h *AlbumHandler) GetAlbumByID(c *gin.Context) {
	id := c.Param("id")
	album, err := h.store.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if album == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
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

	// Se o embedding não foi enviado pelo cliente e temos o serviço Gemini ativo, geramos automaticamente
	if len(album.Embedding.Slice()) == 0 && h.geminiService != nil {
		textToEmbed := fmt.Sprintf("Título: %s. Artista: %s. Descrição: %s", album.Title, album.Artist, album.Description)
		emb, err := h.geminiService.GenerateEmbedding(c.Request.Context(), textToEmbed)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("erro ao gerar embedding: %v", err)})
			return
		}
		album.Embedding = emb
	}

	if err := h.store.Create(c.Request.Context(), album); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, album)
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

	if h.geminiService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "serviço Gemini não está configurado"})
		return
	}

	queryVector, err := h.geminiService.GenerateEmbedding(c.Request.Context(), req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("erro ao gerar embedding da consulta: %v", err)})
		return
	}

	results, err := h.store.SearchSimilar(c.Request.Context(), queryVector, req.Limit, req.MinSimilarity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if results == nil {
		results = []models.AlbumSearchResult{}
	}

	c.JSON(http.StatusOK, results)
}
