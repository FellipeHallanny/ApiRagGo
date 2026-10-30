package services

import (
	"Api_Rag/models"
	"Api_Rag/services/gemini"
	"Api_Rag/store"
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound      = errors.New("álbum não encontrado")
	ErrEmptyTitle    = errors.New("título é obrigatório")
	ErrEmptyArtist   = errors.New("artista é obrigatório")
	ErrNegativePrice = errors.New("preço não pode ser negativo")
	ErrEmptyQuery    = errors.New("a query de busca não pode ser vazia")
)

type AlbumService struct {
	store         *store.AlbumStore
	geminiService *gemini.EmbeddingService
}

func NewAlbumService(store *store.AlbumStore, geminiService *gemini.EmbeddingService) *AlbumService {
	return &AlbumService{
		store:         store,
		geminiService: geminiService,
	}
}

func (s *AlbumService) GetAll(ctx context.Context) ([]models.Album, error) {
	albums, err := s.store.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro no serviço ao buscar álbuns: %w", err)
	}
	if albums == nil {
		return []models.Album{}, nil
	}
	return albums, nil
}

func (s *AlbumService) GetByID(ctx context.Context, id string) (*models.Album, error) {
	if id == "" {
		return nil, errors.New("id do álbum não pode ser vazio")
	}
	album, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("erro no serviço ao buscar álbum por id: %w", err)
	}
	if album == nil {
		return nil, ErrNotFound
	}
	return album, nil
}

func (s *AlbumService) Create(ctx context.Context, album models.Album) (*models.Album, error) {
	if album.Title == "" {
		return nil, ErrEmptyTitle
	}
	if album.Artist == "" {
		return nil, ErrEmptyArtist
	}
	if album.Price < 0 {
		return nil, ErrNegativePrice
	}

	// Se o embedding não foi fornecido e temos o serviço Gemini ativo, geramos automaticamente
	if len(album.Embedding.Slice()) == 0 && s.geminiService != nil {
		textToEmbed := fmt.Sprintf("Título: %s. Artista: %s. Descrição: %s", album.Title, album.Artist, album.Description)
		emb, err := s.geminiService.GenerateEmbedding(ctx, textToEmbed)
		if err != nil {
			return nil, fmt.Errorf("erro ao gerar embedding do álbum: %w", err)
		}
		album.Embedding = emb
	}

	if err := s.store.Create(ctx, album); err != nil {
		return nil, fmt.Errorf("erro no serviço ao cadastrar álbum: %w", err)
	}

	return &album, nil
}

func (s *AlbumService) SearchSimilar(ctx context.Context, query string, limit int, minSimilarity float64) ([]models.AlbumSearchResult, error) {
	if query == "" {
		return nil, ErrEmptyQuery
	}
	if s.geminiService == nil {
		return nil, errors.New("serviço Gemini não está configurado")
	}

	queryVector, err := s.geminiService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar embedding da consulta via Gemini: %w", err)
	}

	results, err := s.store.SearchSimilar(ctx, queryVector, limit, minSimilarity)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar álbuns similares: %w", err)
	}

	if results == nil {
		return []models.AlbumSearchResult{}, nil
	}

	return results, nil
}
