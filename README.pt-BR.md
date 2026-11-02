<div align="center">

# ApiRagGo

*API multi-transporte de alta performance para busca semântica e RAG construída com Go, PostgreSQL (pgvector) e Google Gemini.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![gRPC](https://img.shields.io/badge/gRPC-v1.68-244c5a?style=flat-square&logo=grpc)](https://grpc.io)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12-00ADD8?style=flat-square)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%20%7C%20pgvector-336791?style=flat-square&logo=postgresql)](https://github.com/pgvector/pgvector)
[![Google Gemini](https://img.shields.io/badge/Gemini-text--embedding--004-4285F4?style=flat-square&logo=google)](https://ai.google.dev/)
[![Model Context Protocol](https://img.shields.io/badge/MCP-Protocol-purple?style=flat-square)](https://modelcontextprotocol.io)

**[English](README.md)** • **Português (Brasil)**

[Funcionalidades](#funcionalidades) • [Arquitetura](#arquitetura) • [Pré-requisitos](#pré-requisitos) • [Configuração](#configuração) • [Como Começar](#como-começar) • [Referência da API & Protocolos](#referência-da-api--protocolos) • [Testes](#testes) • [Estrutura do Projeto](#estrutura-do-projeto)

</div>

---

**ApiRagGo** é um backend multi-protocolo de alta performance desenvolvido em Go para busca semântica e recuperação vetorial no padrão RAG (Retrieval-Augmented Generation) sobre catálogos de álbuns musicais. O projeto utiliza o **PostgreSQL com a extensão pgvector** para busca por vizinhos mais próximos (ANN) com índice HNSW, integrando-se ao **Google Gemini** para geração de embeddings vetoriais de 768 dimensões (`text-embedding-004`).

A aplicação disponibiliza três interfaces operacionais conectadas a uma camada unificada de serviços de domínio:
- **HTTP / REST**: API rápida em JSON construída com o framework Gin.
- **gRPC**: Comunicação binária de baixa latência com chamadas unárias, streaming, reflection, interceptors estruturados e health checks.
- **Model Context Protocol (MCP)**: Integração nativa para agentes de inteligência artificial (Claude Desktop, Cursor, JetBrains AI) com ferramentas, recursos e prompts contextuais.

## Funcionalidades

- **Arquitetura Multi-Transporte**: Acesse as funcionalidades de domínio através de REST (`:8080`), gRPC (`:50051`) ou Model Context Protocol (STDIO).
- **Busca Semântica & Recuperação Vetorial**: Realize consultas em linguagem natural calculando a distância de cosseno diretamente no PostgreSQL.
- **Pipeline Automático de Embeddings**: Ao cadastrar itens apenas com texto descritivo, a API gera e indexa automaticamente o vetor de 768 dimensões via Google Gemini (`text-embedding-004`).
- **Indexação HNSW de Baixa Latência**: Utiliza o índice Hierarchical Navigable Small World (`vector_cosine_ops`) do pgvector para buscas na ordem de sub-milissegundos.
- **Descoberta Dinâmica de Modelos**: Identifica dinamicamente os modelos de embedding suportados pela sua chave Gemini, com normalização de dimensionalidade.
- **Streaming em gRPC**: Transmissão em tempo real de resultados de busca via `SearchSimilarStream`, reduzindo o tempo para o primeiro token no cliente.
- **Servidor MCP Completo**: Fornece ferramentas (`search_albums_semantic`, `create_album`, `list_albums`, `get_album_by_id`), recursos (`albums://catalog`) e prompts (`curate_music_by_mood`).
- **Observabilidade Pronta para Produção**: Logs estruturados em formato JSON (`slog`), recuperação contra pânicos, sondas de saúde padrão (`grpc_health_v1`) e gRPC Server Reflection ativado.
- **Ambiente Conteinerizado**: `docker-compose.yml` pré-configurado com PostgreSQL 16, pgvector e inicialização automática via `init.sql`.

## Arquitetura

O diagrama abaixo ilustra a separação entre camadas de transporte, serviços de domínio e armazenamento vetorial:

```mermaid
flowchart TD
    subgraph Clientes["Clientes & Assistentes de IA"]
        C_REST[Aplicações Web / Mobile<br/>HTTP REST]
        C_GRPC[Microsserviços / Evans / grpcurl<br/>gRPC Protobuf]
        C_MCP[Claude Desktop / Cursor / IDEs<br/>Cliente MCP]
    end

    subgraph Transportes["Camada de Transporte"]
        T_REST[Motor Gin<br/>:8080]
        T_GRPC[Servidor gRPC<br/>:50051]
        T_MCP[Servidor MCP<br/>JSON-RPC via STDIO]
    end

    subgraph Nucleo["Serviço de Domínio"]
        SVC[AlbumService<br/>Regras de Negócio e Orquestração]
    end

    subgraph Externo["Armazenamento & IA"]
        GEMINI[Serviço Gemini Embeddings<br/>text-embedding-004]
        STORE[AlbumStore<br/>Repositório SQL pgvector]
        DB[(PostgreSQL 16 + pgvector<br/>Índice HNSW por Cosseno)]
    end

    C_REST -->|HTTP / JSON| T_REST
    C_GRPC -->|Protobuf v3| T_GRPC
    C_MCP -->|JSON-RPC 2.0| T_MCP

    T_REST --> SVC
    T_GRPC --> SVC
    T_MCP --> SVC

    SVC -->|Gera Embeddings| GEMINI
    SVC -->|Persistência & Busca| STORE
    STORE -->|Distância de Cosseno <=> | DB
```

## Pré-requisitos

- [Go](https://go.dev/dl/) (versão 1.22 ou superior)
- [Docker](https://www.docker.com/products/docker-desktop/) e Docker Compose
- Chave de API do [Google AI Studio (Gemini)](https://aistudio.google.com/)
- *(Opcional)* [Buf CLI](https://buf.build/docs/installation) (para regeneração de stubs Protobuf)
- *(Opcional)* [grpcurl](https://github.com/fullstorydev/grpcurl) ou [Evans](https://github.com/ktr0731/evans) (para testes gRPC)

## Configuração

A aplicação é configurada por variáveis de ambiente:

| Variável | Descrição | Valor Padrão | Obrigatória |
|---|---|---|---|
| `GEMINI_API_KEY` | Chave de API do Google Gemini para embeddings de 768 dimensões | *Nenhum* | **Sim** (para busca semântica e geração automática) |
| `DATABASE_URL` | String de conexão com o PostgreSQL | `postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable` | Não |
| `PORT` | Porta de escuta do servidor HTTP REST | `8080` | Não |
| `GRPC_PORT` | Porta de escuta do servidor gRPC | `50051` | Não |

> [!IMPORTANT]
> Se `GEMINI_API_KEY` não for informada, o servidor iniciará em modo reduzido: as operações CRUD convencionais e inserções com vetores manuais funcionarão normalmente, mas a geração automática e as buscas semânticas retornarão erro até a chave ser definida.

## Como Começar

### 1. Clonar o Repositório

```bash
git clone https://github.com/FellipeHallanny/ApiRagGo.git
cd ApiRagGo
```

### 2. Iniciar o PostgreSQL com pgvector

Inicie o container do banco de dados. O script [`init.sql`](file:///init.sql) ativa o `pgvector`, cria a tabela `albums` e gera o índice HNSW no primeiro boot:

```bash
docker compose up -d
```

### 3. Configurar as Variáveis de Ambiente

Defina sua chave de API do Gemini no seu terminal:

**macOS / Linux:**
```bash
export GEMINI_API_KEY="sua-chave-gemini-aqui"
```

**PowerShell (Windows):**
```powershell
$env:GEMINI_API_KEY="sua-chave-gemini-aqui"
```

**Prompt de Comando (CMD - Windows):**
```cmd
set GEMINI_API_KEY=sua-chave-gemini-aqui
```

### 4. Iniciar os Servidores (REST + gRPC)

Inicie o executável que hospeda simultaneamente as portas HTTP e gRPC:

```bash
go mod tidy
go run cmd/api/main.go
```

O log indicará a inicialização dos dois serviços:
```text
Conexão com PostgreSQL (pgvector) estabelecida!
Serviço Gemini (text-embedding-004, 768 dimensões) inicializado!
🚀 Servidor gRPC escutando em :50051 (Reflection e Health ativos)
🌐 Servidor HTTP REST (Gin) escutando em :8080
```

> [!TIP]
> O encerramento dos servidores é gracioso (graceful shutdown). Ao pressionar `Ctrl+C`, todas as conexões ativas terminam ordenadamente antes do processo ser finalizado.

---

## Referência da API & Protocolos

### 1. API HTTP REST (`:8080`)

#### Listar Todos os Álbuns
```bash
curl -X GET http://localhost:8080/albums
```

#### Obter Álbum por ID
```bash
curl -X GET http://localhost:8080/albums/1
```

#### Cadastrar Álbum
Se o campo `embedding` for omitido, o vetor de 768 dimensões é gerado automaticamente pelo Gemini a partir de `title`, `artist` e `description`.

```bash
curl -X POST http://localhost:8080/albums \
  -H "Content-Type: application/json" \
  -d '{
    "id": "1",
    "title": "Blue Train",
    "artist": "John Coltrane",
    "price": 56.99,
    "description": "Álbum seminal de hard bop jazz lançado pela Blue Note em 1957."
  }'
```

#### Busca Semântica (Recuperação RAG)
Pesquise usando descrições em linguagem natural:

```bash
curl -X POST http://localhost:8080/albums/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "discos clássicos de jazz acústico dos anos 50 com saxofone",
    "limit": 3,
    "min_similarity": 0.65
  }'
```

---

### 2. API gRPC (`:50051`)

O serviço gRPC está definido em [`proto/rag/v1/album.proto`](file:///proto/rag/v1/album.proto) sob o pacote `rag.v1.AlbumService`.

#### Definição do Serviço
```protobuf
service AlbumService {
  rpc GetAlbums (GetAlbumsRequest) returns (GetAlbumsResponse);
  rpc GetAlbumByID (GetAlbumByIDRequest) returns (GetAlbumByIDResponse);
  rpc CreateAlbum (CreateAlbumRequest) returns (CreateAlbumResponse);
  rpc SearchSimilar (SearchSimilarRequest) returns (SearchSimilarResponse);
  rpc SearchSimilarStream (SearchSimilarRequest) returns (stream SearchSimilarChunk);
}
```

#### Inspecionar via `grpcurl` (Reflection Habilitado)

Listar serviços disponíveis:
```bash
grpcurl -plaintext localhost:50051 list
```

Descrever os métodos do `AlbumService`:
```bash
grpcurl -plaintext localhost:50051 describe rag.v1.AlbumService
```

#### Executar Busca Semântica via gRPC
```bash
grpcurl -plaintext -d '{
  "query": "álbuns de rock progressivo e guitarras psicodélicas",
  "limit": 3,
  "min_similarity": 0.60
}' localhost:50051 rag.v1.AlbumService/SearchSimilar
```

#### Busca com Streaming de Resultados
```bash
grpcurl -plaintext -d '{
  "query": "improvisações de saxofone no jazz",
  "limit": 5,
  "min_similarity": 0.50
}' localhost:50051 rag.v1.AlbumService/SearchSimilarStream
```

#### Verificar Sonda de Saúde
```bash
grpcurl -plaintext -d '{"service": ""}' localhost:50051 grpc.health.v1.Health/Check
```

---

### 3. Servidor Model Context Protocol (MCP)

O projeto inclui um executável MCP dedicado que se comunica por `STDIO` via JSON-RPC 2.0. Isso possibilita que assistentes como Claude Desktop, Cursor e extensões de IDE consultem o catálogo diretamente.

#### Compilar o Executável MCP

```bash
go build -o bin/mcp.exe ./cmd/mcp
```

#### Configuração de Clientes MCP

Adicione a configuração do ApiRagGo no seu `claude_desktop_config.json` ou nas configurações do Cursor:

```json
{
  "mcpServers": {
    "apirag": {
      "command": "C:\\Projetos\\Go\\ApiRagGo\\bin\\mcp.exe",
      "args": [],
      "env": {
        "DATABASE_URL": "postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable",
        "GEMINI_API_KEY": "sua-chave-gemini-aqui"
      }
    }
  }
}
```

> [!NOTE]
> Todos os logs e diagnósticos operacionais do servidor MCP são direcionados para `stderr`, preservando o canal `stdout` exclusivamente para o tráfego JSON-RPC do protocolo.

#### Capacidades MCP Registradas

| Tipo | Nome / URI | Descrição |
|---|---|---|
| **Tool** | `search_albums_semantic` | Executa busca semântica em linguagem natural com pgvector e Gemini. Argumentos: `query` (obrigatório), `limit` (padrão: 5), `min_similarity` (padrão: 0.60). |
| **Tool** | `get_album_by_id` | Obtém metadados completos de um álbum pelo seu identificador único. Argumento: `id` (obrigatório). |
| **Tool** | `list_albums` | Lista todos os álbuns cadastrados no catálogo. |
| **Tool** | `create_album` | Cadastra novo álbum e gera automaticamente o vetor de 768 dimensões com Gemini. Argumentos: `title`, `artist`, `price`, `description`, `id`. |
| **Resource** | `albums://catalog` | Recurso dinâmico em JSON contendo o catálogo completo de discos. |
| **Resource Template** | `albums://{id}` | Template de URI dinâmico para leitura direta dos detalhes de um álbum por ID. |
| **Prompt** | `curate_music_by_mood` | Orienta o modelo de IA a atuar como curador musical, acionando a busca semântica e gerando recomendações fundamentadas. Argumentos: `mood` (obrigatório), `genre` (opcional). |

---

## Testes

### Testes Automatizados (gRPC & MCP)

Execute a suíte de testes unitários e de integração das camadas de transporte:

```bash
go test -v ./transport/...
```

### Validação com Embeddings Sintéticos

Valide o ranqueamento e o filtro de similaridade vetorial sem consumir cotas da API remota:

```bash
go run cmd/test_embeddings/main.go
```

### Diagnóstico Local do Protocolo MCP

Teste a conformidade do protocolo MCP, o handshake inicial, invocação de ferramentas, leitura de recursos e obtenção de templates de prompt:

```bash
go run cmd/test_mcp/main.go
```

### Regeneração de Código Protobuf

Caso modifique o arquivo [`proto/rag/v1/album.proto`](file:///proto/rag/v1/album.proto), regenere as estruturas Go utilizando o Buf:

```bash
buf generate
```

---

## Estrutura do Projeto

```text
ApiRagGo/
├── app/
│   └── app.go                      # Configuração do router Gin e injeção de dependências
├── cmd/
│   ├── api/
│   │   └── main.go                 # Ponto de entrada unificado (Dual-stack HTTP + gRPC)
│   ├── mcp/
│   │   └── main.go                 # Ponto de entrada do servidor MCP (STDIO JSON-RPC)
│   ├── test_embeddings/
│   │   └── main.go                 # Teste de agrupamento com vetores sintéticos
│   └── test_mcp/
│       └── main.go                 # Cliente de teste automatizado para o servidor MCP
├── db/
│   └── db.go                       # Inicialização do pool de conexões do PostgreSQL
├── gen/
│   └── rag/v1/                     # Código Protobuf e gRPC gerado
│       ├── album.pb.go
│       └── album_grpc.pb.go
├── handlers/
│   └── album_handler.go            # Handlers HTTP REST e validação de payloads
├── models/
│   └── album.go                    # Modelos de domínio e tipos de busca
├── proto/
│   └── rag/v1/
│       └── album.proto             # Definição do serviço e mensagens Protobuf
├── routes/
│   └── routes.go                   # Mapeamento de rotas HTTP
├── services/
│   ├── album_service.go            # Regras de negócio de domínio compartilhadas
│   └── gemini/
│       └── embedding_service.go    # Integração com a API Google Gemini (text-embedding-004)
├── store/
│   └── album_store.go              # Repositório SQL com pgvector e busca HNSW
├── transport/
│   ├── grpc/                       # Implementação da camada gRPC
│   │   ├── album_server.go         # Implementação dos métodos RPC e streaming
│   │   ├── server.go               # Configuração do servidor com logging, health e reflection
│   │   └── server_test.go          # Testes automatizados do servidor gRPC
│   └── mcp/                        # Implementação do Model Context Protocol
│       ├── prompts.go              # Registro de prompts contextuais MCP
│       ├── resources.go            # Registro de recursos e templates de URI
│       ├── server.go               # Configuração e loop STDIO do servidor MCP
│       ├── server_test.go          # Testes automatizados do servidor MCP
│       └── tools.go                # Registro das ferramentas MCP
├── buf.gen.yaml                    # Configuração de plugins de geração do Buf
├── buf.yaml                        # Configuração do módulo Buf
├── docker-compose.yml              # Definição do container PostgreSQL 16 com pgvector
├── go.mod                          # Módulo e dependências Go
├── go.sum                          # Checksums de integridade
├── init.sql                        # Script SQL de inicialização da extensão e índice HNSW
├── mcp-config.sample.json          # Exemplo de configuração para Claude Desktop e Cursor
├── README.md                       # Documentação principal (Inglês)
└── README.pt-BR.md                 # Documentação em Português (Brasil)
```
