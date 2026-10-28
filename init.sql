-- Habilitar a extensão pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Criação da tabela de álbuns com vetor de 768 dimensões
CREATE TABLE IF NOT EXISTS albums (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    description TEXT,
    embedding vector(768),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índice HNSW para busca por similaridade de cosseno (<=>)
CREATE INDEX IF NOT EXISTS idx_albums_embedding_hnsw 
ON albums USING hnsw (embedding vector_cosine_ops);
