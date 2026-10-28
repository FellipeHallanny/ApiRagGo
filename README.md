<div align="center">

# ApiRagGo

*High-performance semantic retrieval and RAG API built with Go, Gin, PostgreSQL (pgvector), and Google Gemini.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12-00ADD8?style=flat-square)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%20%7C%20pgvector-336791?style=flat-square&logo=postgresql)](https://github.com/pgvector/pgvector)
[![Google Gemini](https://img.shields.io/badge/Gemini-text--embedding--004-4285F4?style=flat-square&logo=google)](https://ai.google.dev/)

**English** • **[Português (Brasil)](README.pt-BR.md)**

[Features](#features) • [Architecture](#architecture) • [Prerequisites](#prerequisites) • [Configuration](#configuration) • [Getting Started](#getting-started) • [API Reference](#api-reference) • [Testing](#testing)

</div>

---

**ApiRagGo** is a lightweight, production-ready REST API implementing semantic search and Retrieval-Augmented Generation (RAG) capabilities over music album catalogs. Built with Go and the Gin framework, it leverages **PostgreSQL with pgvector** for fast Approximate Nearest Neighbor (ANN) search via HNSW indexing and integrates seamlessly with the **Google Gemini Generative Language API** for generating 768-dimensional text embeddings.

## Features

- **Semantic Search & Vector Retrieval**: Query your catalog using natural language. Matches are computed using cosine similarity directly inside PostgreSQL.
- **Automatic Embedding Pipeline**: Ingest new records with simple text metadata; the API automatically generates and indexes 768-dimension vectors via Google Gemini (`text-embedding-004`).
- **Dynamic Model Discovery**: Automatically inspects available Gemini embedding models for your API key with built-in dimension normalization and fallback mechanisms.
- **High-Performance HNSW Indexing**: Uses pgvector's Hierarchical Navigable Small World (`vector_cosine_ops`) index for sub-millisecond similarity lookups.
- **Similarity Threshold Filtering**: Built-in support for confidence scores (`min_similarity`) to prevent false positives and irrelevant search results.
- **Clean Architecture**: Decoupled domain models, persistence store, HTTP transport handlers, and external AI service layers.
- **Containerized Database**: Pre-configured `docker-compose.yml` bundling PostgreSQL 16, pgvector, and automatic schema initialization.

## Architecture

The diagram below illustrates the ingestion and retrieval workflows:

```mermaid
flowchart TD
    subgraph Ingestion["1. Document Ingestion (POST /albums)"]
        A[Client Request] --> B[Gin Handler]
        B --> C{Embedding Provided?}
        C -->|No| D[Gemini Service<br/>text-embedding-004]
        D -->|768d Vector| E[PostgreSQL Store]
        C -->|Yes| E
        E --> F[(PostgreSQL + pgvector<br/>HNSW Index)]
    end

    subgraph Retrieval["2. Semantic Search (POST /albums/search)"]
        G[Natural Language Query] --> H[Gin Handler]
        H --> I[Gemini Service<br/>text-embedding-004]
        I -->|Query Vector| J[PostgreSQL Store]
        J -->|Cosine Distance <=> | F
        F -->|Filtered by min_similarity| K[Ranked Results + Score]
        K --> L[Client Response]
    end
```

## Prerequisites

Before running the application, ensure you have the following installed:

- [Go](https://go.dev/dl/) (version 1.22 or higher)
- [Docker](https://www.docker.com/products/docker-desktop/) and Docker Compose
- A [Google Gemini API Key](https://aistudio.google.com/)

## Configuration

The application is configured through environment variables:

| Variable | Description | Default | Required |
|---|---|---|---|
| `GEMINI_API_KEY` | API key for Google Gemini embedding services | *None* | **Yes** (for semantic search & automatic embeddings) |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable` | No |

> [!IMPORTANT]
> If `GEMINI_API_KEY` is omitted, the API will start in a degraded mode. CRUD endpoints and manual vector insertions will work, but automatic vector generation and `/albums/search` will return an error until the key is provided.

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/FellipeHallanny/ApiRagGo.git
cd ApiRagGo
```

### 2. Start PostgreSQL with pgvector

Launch the database container using Docker Compose. The `init.sql` script creates the table and HNSW index on first startup:

```bash
docker compose up -d
```

To verify the container is healthy:

```bash
docker compose ps
```

### 3. Set the Environment Variables

Set your Gemini API key in your current shell:

**Bash / macOS / Linux:**
```bash
export GEMINI_API_KEY="your-gemini-api-key"
```

**PowerShell (Windows):**
```powershell
$env:GEMINI_API_KEY="your-gemini-api-key"
```

**Command Prompt (Windows):**
```cmd
set GEMINI_API_KEY=your-gemini-api-key
```

### 4. Run the API Server

Download dependencies and start the HTTP server:

```bash
go mod tidy
go run cmd/api/main.go
```

The server will listen at `http://localhost:8080`.

> [!TIP]
> The database schema is pre-configured with `init.sql`. If you are using an external PostgreSQL instance, execute the contents of [`init.sql`](file:///init.sql) to enable the `vector` extension and create the required tables.

## API Reference

### 1. List All Albums

Retrieve all albums currently in the database.

- **URL**: `GET /albums`
- **Response**: `200 OK`

```bash
curl -X GET http://localhost:8080/albums
```

```json
[
  {
    "id": "1",
    "title": "Blue Train",
    "artist": "John Coltrane",
    "price": 56.99,
    "description": "Seminal hard bop jazz album released by Blue Note in 1957."
  }
]
```

---

### 2. Get Album by ID

Fetch a single album by its unique identifier.

- **URL**: `GET /albums/:id`
- **Response**: `200 OK` or `404 Not Found`

```bash
curl -X GET http://localhost:8080/albums/1
```

---

### 3. Create an Album

Create an album record. If the `embedding` field is omitted, the service will generate a 768-dimensional vector automatically using Gemini based on the title, artist, and description.

- **URL**: `POST /albums`
- **Content-Type**: `application/json`
- **Response**: `201 Created`

```bash
curl -X POST http://localhost:8080/albums \
  -H "Content-Type: application/json" \
  -d '{
    "id": "1",
    "title": "Blue Train",
    "artist": "John Coltrane",
    "price": 56.99,
    "description": "Seminal hard bop jazz album released by Blue Note in 1957."
  }'
```

> [!NOTE]
> When providing an explicit `embedding`, supply an array of 768 floating point numbers matching the vector schema.

---

### 4. Semantic Search (RAG Retrieval)

Perform a vector similarity search across all stored albums using natural language queries.

- **URL**: `POST /albums/search`
- **Content-Type**: `application/json`
- **Response**: `200 OK`

#### Request Parameters

| Field | Type | Description | Default |
|---|---|---|---|
| `query` | string | Natural language question or search phrase | *Required* |
| `limit` | integer | Maximum number of results to return | `5` |
| `min_similarity` | float | Minimum cosine similarity threshold (`0.0` to `1.0`) | `0.60` |

#### Example Request

```bash
curl -X POST http://localhost:8080/albums/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "classic 50s acoustic jazz saxophone and trumpet records",
    "limit": 3,
    "min_similarity": 0.65
  }'
```

#### Example Response

```json
[
  {
    "id": "1",
    "title": "Blue Train",
    "artist": "John Coltrane",
    "price": 56.99,
    "description": "Seminal hard bop jazz album released by Blue Note in 1957.",
    "similarity": 0.8241
  }
]
```

## Testing

A standalone test utility is included under `cmd/test_embeddings` to validate vector proximity and similarity filtering using synthetic embeddings without making remote API calls:

1. Ensure the API server is running on `localhost:8080`.
2. In a separate terminal, execute:

```bash
go run cmd/test_embeddings/main.go
```

The script will:
- Register synthetic cluster vectors (e.g., Jazz records clustered near ~0.80 and Rock records near ~0.10).
- Run a similarity search against the Jazz vector space.
- Verify that similarity ranking and threshold filtering return expected semantic matches.

## Project Structure

```text
ApiRagGo/
├── app/
│   └── app.go                   # Application routing and dependency setup
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entrypoint
│   └── test_embeddings/
│       └── main.go              # Synthetic vector test utility
├── db/
│   └── db.go                    # PostgreSQL connection pool initializer
├── handlers/
│   └── album_handler.go         # HTTP handlers and request validation
├── models/
│   └── album.go                 # Data models and search response types
├── routes/
│   └── routes.go                # Route definitions
├── services/
│   └── gemini/
│       └── embedding_service.go # Google Gemini API integration and model discovery
├── store/
│   └── album_store.go           # pgvector SQL repository and similarity queries
├── docker-compose.yml           # PostgreSQL 16 + pgvector container definition
├── go.mod                       # Go module dependencies
├── go.sum                       # Go module checksums
├── init.sql                     # Database schema and HNSW index definition
├── README.md                    # Project documentation (English)
└── README.pt-BR.md              # Project documentation (Portuguese)
```
