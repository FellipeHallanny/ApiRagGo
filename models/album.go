package models

import "github.com/pgvector/pgvector-go"

type Album struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Artist      string          `json:"artist"`
	Price       float64         `json:"price"`
	Description string          `json:"description,omitempty"`
	Embedding   pgvector.Vector `json:"embedding,omitempty"`
}

type AlbumSearchResult struct {
	Album
	Similarity float64 `json:"similarity"`
}
