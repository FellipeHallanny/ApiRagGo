<div align="center">

# ApiRagGo

*High-performance multi-transport semantic retrieval and RAG API built with Go, PostgreSQL (pgvector), and Google Gemini.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-v1.68-244c5a?style=flat-square&logo=grpc)](https://grpc.io)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12-00ADD8?style=flat-square)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%20%7C%20pgvector-336791?style=flat-square&logo=postgresql)](https://github.com/pgvector/pgvector)
[![Google Gemini](https://img.shields.io/badge/Gemini-text--embedding--004-4285F4?style=flat-square&logo=google)](https://ai.google.dev/)
[![Model Context Protocol](https://img.shields.io/badge/MCP-Protocol-purple?style=flat-square)](https://modelcontextprotocol.io)

**English** • **[Português (Brasil)](README.pt-BR.md)**

[Features](#features) • [Architecture](#architecture) • [Prerequisites](#prerequisites) • [Configuration](#configuration) • [Getting Started](#getting-started) • [API & Protocol Reference](#api--protocol-reference) • [Testing](#testing) • [Project Structure](#project-structure)

</div>

---

**ApiRagGo** is a multi-protocol semantic search and Retrieval-Augmented Generation (RAG) backend engineered in Go. It indexes music album catalogs using **PostgreSQL with pgvector** for sub-millisecond Approximate Nearest Neighbor (ANN) search via HNSW indexing, powered by **Google Gemini** 768-dimensional embeddings (`text-embedding-004`).

The application exposes three concurrent interfaces powered by a unified domain service layer:
- **HTTP / REST**: Fast JSON API built with Gin.
- **gRPC**: Low-latency RPC with unary & server streaming, reflection, structured interceptors, and health checks.
- **Model Context Protocol (MCP)**: Native integration for AI clients (Claude Desktop, Cursor, JetBrains AI) with tools, resources, and contextual prompts.

## Features

- **Multi-Transport Architecture**: Access domain capabilities interchangeably via REST (`:8080`), gRPC (`:50051`), or Model Context Protocol (STDIO).
- **Semantic Search & Vector Retrieval**: Query records using natural language. Matches are ranked using cosine distance directly in PostgreSQL.
- **Automated Embedding Pipeline**: Ingest catalog items with standard textual descriptions; the service generates 768-dimensional embeddings via Google Gemini (`text-embedding-004`).
- **High-Performance HNSW Indexing**: Uses pgvector's Hierarchical Navigable Small World index (`vector_cosine_ops`) for real-time similarity lookups.
- **Dynamic Model Discovery**: Inspects Gemini embedding models dynamically with vector dimension normalization and graceful fallbacks.
- **Streaming gRPC Endpoints**: Stream similarity search results chunk-by-chunk to reduce client time-to-first-token.
- **Model Context Protocol (MCP) Server**: Provides native MCP tools (`search_albums_semantic`, `create_album`, `list_albums`, `get_album_by_id`), resources (`albums://catalog`), and prompts (`curate_music_by_mood`).
- **Production-Ready Observability**: Structured JSON logging (`slog`), gRPC panic recovery, standard health check probe (`grpc_health_v1`), and gRPC server reflection.
- **Containerized Database**: Pre-configured `docker-compose.yml` with PostgreSQL 16, pgvector, and automated schema migration.

## Architecture

The following diagram illustrates the decoupled architecture across transport layers, domain services, and vector persistence:

```mermaid
flowchart TD
    subgraph Clients["Clients & AI Assistants"]
        C_REST[Web / Mobile Clients<br/>HTTP REST]
        C_GRPC[Microservices / Evans / grpcurl<br/>gRPC Protobuf]
        C_MCP[Claude Desktop / Cursor / IDEs<br/>MCP Client]
    end

    subgraph Transports["Transport Layer"]
        T_REST[Gin Engine<br/>:8080]
        T_GRPC[gRPC Server<br/>:50051]
        T_MCP[MCP Server<br/>JSON-RPC via STDIO]
    end

    subgraph Core["Domain Service"]
        SVC[AlbumService<br/>Domain Business Logic]
    end

    subgraph External["AI & Storage"]
        GEMINI[Gemini Embedding Service<br/>text-embedding-004]
        STORE[AlbumStore<br/>pgvector SQL Repository]
        DB[(PostgreSQL 16 + pgvector<br/>HNSW Cosine Index)]
    end

    C_REST -->|HTTP / JSON| T_REST
    C_GRPC -->|Protobuf v3| T_GRPC
    C_MCP -->|JSON-RPC 2.0| T_MCP

    T_REST --> SVC
    T_GRPC --> SVC
    T_MCP --> SVC

    SVC -->|Generate Embeddings| GEMINI
    SVC -->|Persist & Query| STORE
    STORE -->|Cosine Distance <=> | DB
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.22 or higher
- [Docker](https://www.docker.com/products/docker-desktop/) with Docker Compose
- [Google Gemini API Key](https://aistudio.google.com/)
- *(Optional)* [Buf CLI](https://buf.build/docs/installation) (for regenerating protobuf stubs)
- *(Optional)* [grpcurl](https://github.com/fullstorydev/grpcurl) or [Evans](https://github.com/ktr0731/evans) (for interactive gRPC testing)

## Configuration

The application is configured using environment variables:

| Variable | Description | Default | Required |
|---|---|---|---|
| `GEMINI_API_KEY` | Google Gemini API key for generating 768-dimension embeddings | *None* | **Yes** (for semantic search & automatic embeddings) |
| `DATABASE_URL` | PostgreSQL connection URL | `postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable` | No |
| `PORT` | HTTP REST server port | `8080` | No |
| `GRPC_PORT` | gRPC server port | `50051` | No |

> [!IMPORTANT]
> If `GEMINI_API_KEY` is not provided, the API starts in degraded mode. Standard CRUD operations and manual embedding insertions will function normally, while automatic embedding generation and semantic search queries will return an error until the key is set.

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/FellipeHallanny/ApiRagGo.git
cd ApiRagGo
```

### 2. Start PostgreSQL with pgvector

Launch the database container. The included [`init.sql`](file:///init.sql) script enables `pgvector`, creates the `albums` table, and builds the HNSW index on initial boot:

```bash
docker compose up -d
```

### 3. Configure the Environment

Set your Gemini API key in your terminal session:

**macOS / Linux:**
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

### 4. Run the API Server (REST + gRPC)

Start the unified server hosting both REST and gRPC simultaneously:

```bash
go mod tidy
go run cmd/api/main.go
```

The output will confirm both listeners are online:
```text
Conexão com PostgreSQL (pgvector) estabelecida!
Serviço Gemini (text-embedding-004, 768 dimensões) inicializado!
🚀 Servidor gRPC escutando em :50051 (Reflection e Health ativos)
🌐 Servidor HTTP REST (Gin) escutando em :8080
```

> [!TIP]
> The server implements graceful shutdown. Pressing `Ctrl+C` cleanly finishes in-flight requests on both HTTP and gRPC listeners before terminating.

---

## API & Protocol Reference

### 1. REST API (`:8080`)

#### List All Albums
```bash
curl -X GET http://localhost:8080/albums
```

#### Get Album by ID
```bash
curl -X GET http://localhost:8080/albums/1
```

#### Create Album
Embeddings are computed automatically from `title`, `artist`, and `description` when the `embedding` field is omitted.

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

#### Semantic Search (RAG Retrieval)
Perform natural language proximity matching against pgvector:

```bash
curl -X POST http://localhost:8080/albums/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "classic 50s acoustic jazz saxophone and trumpet records",
    "limit": 3,
    "min_similarity": 0.65
  }'
```

---

### 2. gRPC API (`:50051`)

The gRPC service is defined in [`proto/rag/v1/album.proto`](file:///proto/rag/v1/album.proto) under package `rag.v1.AlbumService`.

#### Service Definition
```protobuf
service AlbumService {
  rpc GetAlbums (GetAlbumsRequest) returns (GetAlbumsResponse);
  rpc GetAlbumByID (GetAlbumByIDRequest) returns (GetAlbumByIDResponse);
  rpc CreateAlbum (CreateAlbumRequest) returns (CreateAlbumResponse);
  rpc SearchSimilar (SearchSimilarRequest) returns (SearchSimilarResponse);
  rpc SearchSimilarStream (SearchSimilarRequest) returns (stream SearchSimilarChunk);
}
```

#### Inspect with `grpcurl` (Server Reflection Enabled)

List services:
```bash
grpcurl -plaintext localhost:50051 list
```

Describe `AlbumService`:
```bash
grpcurl -plaintext localhost:50051 describe rag.v1.AlbumService
```

#### Execute Semantic Search via gRPC
```bash
grpcurl -plaintext -d '{
  "query": "psychedelic rock guitar albums with progressive arrangements",
  "limit": 3,
  "min_similarity": 0.60
}' localhost:50051 rag.v1.AlbumService/SearchSimilar
```

#### Stream Semantic Matches
```bash
grpcurl -plaintext -d '{
  "query": "jazz saxophone improvisation",
  "limit": 5,
  "min_similarity": 0.50
}' localhost:50051 rag.v1.AlbumService/SearchSimilarStream
```

#### Check Health Probe
```bash
grpcurl -plaintext -d '{"service": ""}' localhost:50051 grpc.health.v1.Health/Check
```

---

### 3. Model Context Protocol (MCP) Server

ApiRagGo includes a dedicated MCP binary communicating over `STDIO` via JSON-RPC 2.0. This allows AI assistants such as Claude Desktop, Cursor, and IDE extensions to discover and use the album catalog directly.

#### Build the MCP Binary

```bash
go build -o bin/mcp.exe ./cmd/mcp
```

#### MCP Client Configuration

Add ApiRagGo to your Claude Desktop config (`claude_desktop_config.json`) or Cursor MCP settings:

```json
{
  "mcpServers": {
    "apirag": {
      "command": "C:\\Projetos\\Go\\ApiRagGo\\bin\\mcp.exe",
      "args": [],
      "env": {
        "DATABASE_URL": "postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable",
        "GEMINI_API_KEY": "your-gemini-api-key"
      }
    }
  }
}
```

> [!NOTE]
> All operational diagnostics and logging in the MCP server are directed to `stderr`, keeping `stdout` reserved strictly for clean JSON-RPC protocol frames.

#### Available MCP Capabilities

| Type | Name / URI | Description |
|---|---|---|
| **Tool** | `search_albums_semantic` | Executes natural language semantic search using pgvector and Gemini embeddings. Arguments: `query` (required), `limit` (default: 5), `min_similarity` (default: 0.60). |
| **Tool** | `get_album_by_id` | Retrieves full metadata of a specific album by its unique identifier. Argument: `id` (required). |
| **Tool** | `list_albums` | Returns all albums registered in the catalog repository. |
| **Tool** | `create_album` | Ingests a new album and generates a 768-dimensional vector embedding automatically. Arguments: `title`, `artist`, `price`, `description`, `id`. |
| **Resource** | `albums://catalog` | Dynamic JSON resource exposing the entire album catalog. |
| **Resource Template** | `albums://{id}` | Dynamic URI template fetching details for a specific album by ID. |
| **Prompt** | `curate_music_by_mood` | Directs the AI model to act as a music curator, calling `search_albums_semantic` and generating justified recommendations. Arguments: `mood` (required), `genre` (optional). |

---

## Testing

### Unit & Integration Tests

Run the test suite covering the gRPC and MCP transport packages:

```bash
go test -v ./transport/...
```

### Synthetic Vector Benchmark

Test vector clustering, cosine ranking, and threshold filtering without consuming external Gemini API quotas:

```bash
go run cmd/test_embeddings/main.go
```

### MCP Server Local Diagnostic

Validate MCP JSON-RPC protocol compliance, handshake, tool invocation, resource reading, and prompt generation over STDIO:

```bash
go run cmd/test_mcp/main.go
```

### Regenerate Protobuf Code

If you modify [`proto/rag/v1/album.proto`](file:///proto/rag/v1/album.proto), regenerate Go code using Buf:

```bash
buf generate
```

---

## Project Structure

```text
ApiRagGo/
├── app/
│   └── app.go                      # Gin HTTP router setup & dependency injection
├── cmd/
│   ├── api/
│   │   └── main.go                 # Unified API entrypoint (Dual-stack HTTP + gRPC)
│   ├── mcp/
│   │   └── main.go                 # MCP server entrypoint (JSON-RPC over STDIO)
│   ├── test_embeddings/
│   │   └── main.go                 # Synthetic vector clustering test script
│   └── test_mcp/
│       └── main.go                 # Automated MCP client handshake & tool runner
├── db/
│   └── db.go                       # PostgreSQL connection pool initializer
├── gen/
│   └── rag/v1/                     # Generated Protobuf & gRPC code
│       ├── album.pb.go
│       └── album_grpc.pb.go
├── handlers/
│   └── album_handler.go            # REST HTTP handlers & payload validation
├── models/
│   └── album.go                    # Core domain entities & search structures
├── proto/
│   └── rag/v1/
│       └── album.proto             # Protocol Buffers service & message definitions
├── routes/
│   └── routes.go                   # HTTP route registrations
├── services/
│   ├── album_service.go            # Shared domain business logic
│   └── gemini/
│       └── embedding_service.go    # Google Gemini API integration (text-embedding-004)
├── store/
│   └── album_store.go              # pgvector SQL repository & HNSW similarity search
├── transport/
│   ├── grpc/                       # gRPC transport implementation
│   │   ├── album_server.go         # gRPC service implementation & streaming
│   │   ├── server.go               # Server setup with logging, health & reflection
│   │   └── server_test.go          # gRPC integration tests
│   └── mcp/                        # Model Context Protocol transport
│       ├── prompts.go              # MCP prompts registration
│       ├── resources.go            # MCP resources & URI templates
│       ├── server.go               # MCP STDIO server setup
│       ├── server_test.go          # MCP integration tests
│       └── tools.go                # MCP tools registration
├── buf.gen.yaml                    # Buf code generation template
├── buf.yaml                        # Buf workspace configuration
├── docker-compose.yml              # PostgreSQL 16 + pgvector container definition
├── go.mod                          # Go module dependencies
├── go.sum                          # Go module checksums
├── init.sql                        # Database schema, pgvector extension & HNSW index
├── mcp-config.sample.json          # Example configuration for Claude Desktop / Cursor
├── README.md                       # Documentation (English)
└── README.pt-BR.md                 # Documentação (Português)
```
