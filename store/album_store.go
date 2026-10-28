package store

import (
	"Api_Rag/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pgvector/pgvector-go"
)

type AlbumStore struct {
	db *sql.DB
}

// NewAlbumStore cria uma nova instância do repositório de álbuns
func NewAlbumStore(db *sql.DB) *AlbumStore {
	return &AlbumStore{db: db}
}

// GetAll retorna todos os álbuns cadastrados
func (s *AlbumStore) GetAll(ctx context.Context) ([]models.Album, error) {
	query := `SELECT id, title, artist, price, COALESCE(description, '') FROM albums ORDER BY title ASC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar álbuns: %w", err)
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.ID, &a.Title, &a.Artist, &a.Price, &a.Description); err != nil {
			return nil, fmt.Errorf("falha ao ler registro de álbum: %w", err)
		}
		albums = append(albums, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro na iteração dos resultados: %w", err)
	}

	return albums, nil
}

// GetByID busca um álbum pelo identificador
func (s *AlbumStore) GetByID(ctx context.Context, id string) (*models.Album, error) {
	query := `SELECT id, title, artist, price, COALESCE(description, ''), embedding FROM albums WHERE id = $1`
	var a models.Album
	err := s.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.Title, &a.Artist, &a.Price, &a.Description, &a.Embedding)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("falha ao consultar álbum por ID: %w", err)
	}
	return &a, nil
}

// Create insere um novo álbum no PostgreSQL com suporte ao vetor de 768 dimensões
func (s *AlbumStore) Create(ctx context.Context, album models.Album) error {
	query := `
		INSERT INTO albums (id, title, artist, price, description, embedding)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	var embArg any = nil
	if len(album.Embedding.Slice()) > 0 {
		embArg = album.Embedding
	}

	_, err := s.db.ExecContext(ctx, query, album.ID, album.Title, album.Artist, album.Price, album.Description, embArg)
	if err != nil {
		return fmt.Errorf("falha ao inserir álbum: %w", err)
	}
	return nil
}

// SearchSimilar realiza busca semântica filtrando por similaridade mínima (threshold)
func (s *AlbumStore) SearchSimilar(ctx context.Context, queryEmbedding pgvector.Vector, limit int, minSimilarity float64) ([]models.AlbumSearchResult, error) {
	if limit <= 0 {
		limit = 5
	}
	if minSimilarity <= 0 {
		minSimilarity = 0.60 // Padrão seguro para evitar falsos positivos
	}

	query := `
		SELECT id, title, artist, price, COALESCE(description, ''),
		       1 - (embedding <=> $1) AS similarity
		FROM albums
		WHERE embedding IS NOT NULL
		  AND (1 - (embedding <=> $1)) >= $2
		ORDER BY embedding <=> $1
		LIMIT $3;
	`

	rows, err := s.db.QueryContext(ctx, query, queryEmbedding, minSimilarity, limit)
	if err != nil {
		return nil, fmt.Errorf("falha na busca por similaridade: %w", err)
	}
	defer rows.Close()

	var results []models.AlbumSearchResult
	for rows.Next() {
		var res models.AlbumSearchResult
		if err := rows.Scan(&res.ID, &res.Title, &res.Artist, &res.Price, &res.Description, &res.Similarity); err != nil {
			return nil, fmt.Errorf("falha ao ler resultado da busca: %w", err)
		}
		results = append(results, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro na iteração da busca por similaridade: %w", err)
	}

	return results, nil
}
