<div align="center">

# ApiRagGo

*API REST de alta performance para busca semântica e RAG construída com Go, Gin, PostgreSQL (pgvector) e Google Gemini.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.12-00ADD8?style=flat-square)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%20%7C%20pgvector-336791?style=flat-square&logo=postgresql)](https://github.com/pgvector/pgvector)
[![Google Gemini](https://img.shields.io/badge/Gemini-text--embedding--004-4285F4?style=flat-square&logo=google)](https://ai.google.dev/)

**[English](README.md)** • **Português (Brasil)**

[Funcionalidades](#funcionalidades) • [Arquitetura](#arquitetura) • [Pré-requisitos](#pré-requisitos) • [Configuração](#configuração) • [Como Começar](#como-começar) • [Referência da API](#referência-da-api) • [Testes](#testes)

</div>

---

**ApiRagGo** é uma API REST desenvolvida em Go e Gin, projetada para busca semântica e recuperação de informações no padrão RAG (Retrieval-Augmented Generation) sobre catálogos de álbuns musicais. O projeto utiliza o **PostgreSQL com a extensão pgvector** para busca vetorial de alta performance por vizinhos mais próximos (ANN) via índice HNSW, integrando-se à API do **Google Gemini** para geração de embeddings vetoriais de 768 dimensões.

## Funcionalidades

- **Busca Semântica & Recuperação Vetorial**: Permite consultas usando linguagem natural. As correspondências são calculadas por similaridade de cosseno diretamente no PostgreSQL.
- **Pipeline Automático de Embeddings**: Ao cadastrar um álbum contendo apenas título, artista e descrição, a API gera e indexa automaticamente o vetor de 768 dimensões via Google Gemini (`text-embedding-004`).
- **Descoberta Dinâmica de Modelos**: Identifica dinamicamente os modelos de embedding suportados pela sua chave da API Gemini, com tratamento e normalização dimensional.
- **Indexação HNSW de Alta Performance**: Utiliza o índice Hierarchical Navigable Small World (`vector_cosine_ops`) do pgvector para buscas com latência de sub-milissegundos.
- **Filtragem por Limiar de Similaridade**: Parâmetro `min_similarity` configurável para evitar falsos positivos e retornar apenas resultados relevantes.
- **Arquitetura em Camadas (Clean Architecture)**: Separação clara entre modelos de domínio, repositório de dados, handlers HTTP e serviços de inteligência artificial.
- **Ambiente Conteinerizado**: `docker-compose.yml` pronto para uso com PostgreSQL 16, pgvector e inicialização automática do esquema via `init.sql`.

## Arquitetura

O diagrama abaixo descreve o fluxo de ingestão e recuperação de dados:

```mermaid
flowchart TD
    subgraph Ingestao["1. Ingestão de Documentos (POST /albums)"]
        A[Requisição do Cliente] --> B[Handler Gin]
        B --> C{Embedding Informado?}
        C -->|Não| D[Serviço Gemini<br/>text-embedding-004]
        D -->|Vetor 768d| E[Repositório PostgreSQL]
        C -->|Sim| E
        E --> F[(PostgreSQL + pgvector<br/>Índice HNSW)]
    end

    subgraph Recuperacao["2. Busca Semântica (POST /albums/search)"]
        G[Consulta em Texto Natural] --> H[Handler Gin]
        H --> I[Serviço Gemini<br/>text-embedding-004]
        I -->|Vetor da Consulta| J[Repositório PostgreSQL]
        J -->|Distância de Cosseno <=> | F
        F -->|Filtrado por min_similarity| K[Resultados Ranqueados + Score]
        K --> L[Resposta ao Cliente]
    end
```

## Pré-requisitos

Antes de iniciar, certifique-se de ter instalado:

- [Go](https://go.dev/dl/) (versão 1.22 ou superior)
- [Docker](https://www.docker.com/products/docker-desktop/) e Docker Compose
- Uma chave de API do [Google AI Studio (Gemini)](https://aistudio.google.com/)

## Configuração

A aplicação é configurada por meio de variáveis de ambiente:

| Variável | Descrição | Valor Padrão | Obrigatória |
|---|---|---|---|
| `GEMINI_API_KEY` | Chave da API do Google Gemini para geração de embeddings | *Nenhum* | **Sim** (para busca semântica e embeddings automáticos) |
| `DATABASE_URL` | String de conexão com o PostgreSQL | `postgres://postgres:postgrespassword@localhost:5432/apirag?sslmode=disable` | Não |

> [!IMPORTANT]
> Caso a variável `GEMINI_API_KEY` não seja definida, o servidor iniciará em modo reduzido: as operações CRUD convencionais e inserções com vetores manuais funcionarão normalmente, porém a geração automática de embeddings e a rota `/albums/search` retornarão erro até que a chave seja configurada.

## Como Começar

### 1. Clonar o Repositório

```bash
git clone https://github.com/FellipeHallanny/ApiRagGo.git
cd ApiRagGo
```

### 2. Iniciar o PostgreSQL com pgvector

Inicie o container do banco de dados via Docker Compose. O script `init.sql` criará a extensão, a tabela e o índice HNSW automaticamente:

```bash
docker compose up -d
```

Verifique se o container está ativo:

```bash
docker compose ps
```

### 3. Configurar a Chave do Gemini

Defina a variável no terminal antes de iniciar a aplicação:

**Bash / macOS / Linux:**
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

### 4. Executar o Servidor

Instale as dependências e inicie a API:

```bash
go mod tidy
go run cmd/api/main.go
```

A API estará disponível em `http://localhost:8080`.

> [!TIP]
> Caso utilize uma instância externa do PostgreSQL, execute os comandos presentes em [`init.sql`](file:///init.sql) para ativar a extensão `vector` e criar o esquema da tabela.

## Referência da API

### 1. Listar Todos os Álbuns

Retorna a lista com todos os álbuns cadastrados no banco.

- **URL**: `GET /albums`
- **Resposta**: `200 OK`

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
    "description": "Álbum seminal de hard bop jazz lançado pela Blue Note em 1957."
  }
]
```

---

### 2. Obter Álbum por ID

Busca um registro específico pelo seu identificador.

- **URL**: `GET /albums/:id`
- **Resposta**: `200 OK` ou `404 Not Found`

```bash
curl -X GET http://localhost:8080/albums/1
```

---

### 3. Cadastrar Álbum

Insere um novo álbum. Se o campo `embedding` não for informado no payload, o serviço Gemini gerará automaticamente um vetor de 768 dimensões combinando os campos de título, artista e descrição.

- **URL**: `POST /albums`
- **Content-Type**: `application/json`
- **Resposta**: `201 Created`

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

> [!NOTE]
> Para fornecer um vetor pré-calculado manualmente, envie a propriedade `embedding` contendo um array de 768 números float.

---

### 4. Busca Semântica (Recuperação RAG)

Realiza pesquisa semântica por proximidade vetorial a partir de uma consulta em texto natural.

- **URL**: `POST /albums/search`
- **Content-Type**: `application/json`
- **Resposta**: `200 OK`

#### Parâmetros da Requisição

| Campo | Tipo | Descrição | Padrão |
|---|---|---|---|
| `query` | string | Pergunta ou texto de busca em linguagem natural | *Obrigatório* |
| `limit` | integer | Limite máximo de resultados retornados | `5` |
| `min_similarity` | float | Limiar mínimo de similaridade de cosseno (`0.0` a `1.0`) | `0.60` |

#### Exemplo de Requisição

```bash
curl -X POST http://localhost:8080/albums/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "discos clássicos de jazz acústico dos anos 50 com saxofone",
    "limit": 3,
    "min_similarity": 0.65
  }'
```

#### Exemplo de Resposta

```json
[
  {
    "id": "1",
    "title": "Blue Train",
    "artist": "John Coltrane",
    "price": 56.99,
    "description": "Álbum seminal de hard bop jazz lançado pela Blue Note em 1957.",
    "similarity": 0.8241
  }
]
```

## Testes

O projeto inclui um utilitário em `cmd/test_embeddings` para validar a ordenação e filtragem por similaridade utilizando embeddings sintéticos, sem consumir cotas da API remota:

1. Certifique-se de que a API esteja em execução em `localhost:8080`.
2. Em outro terminal, execute:

```bash
go run cmd/test_embeddings/main.go
```

O script executa os seguintes passos:
- Cadastra clusters vetoriais sintéticos (ex.: álbuns de Jazz próximos a ~0.80 e Rock próximos a ~0.10).
- Dispara uma busca semântica no espaço vetorial de Jazz.
- Valida se os resultados são devidamente ordenados por relevância e filtrados pelo limiar estipulado.

## Estrutura do Projeto

```text
ApiRagGo/
├── app/
│   └── app.go                   # Configuração de rotas e injeção de dependências
├── cmd/
│   ├── api/
│   │   └── main.go              # Ponto de entrada do servidor HTTP
│   └── test_embeddings/
│       └── main.go              # Utilitário de testes com vetores sintéticos
├── db/
│   └── db.go                    # Conexão e pool do PostgreSQL
├── handlers/
│   └── album_handler.go         # Handlers HTTP e validação de payloads
├── models/
│   └── album.go                 # Modelos de dados e estruturas de resposta
├── routes/
│   └── routes.go                # Registro e mapeamento de rotas
├── services/
│   └── gemini/
│       └── embedding_service.go # Integração com API REST do Google Gemini
├── store/
│   └── album_store.go           # Repositório SQL com pgvector e busca vetorial
├── docker-compose.yml           # Definição do container PostgreSQL 16 com pgvector
├── go.mod                       # Módulos e dependências Go
├── go.sum                       # Checksums de integridade do Go
├── init.sql                     # Script de criação da extensão pgvector e tabelas
├── README.md                    # Documentação do projeto (Inglês)
└── README.pt-BR.md              # Documentação do projeto (Português)
```
