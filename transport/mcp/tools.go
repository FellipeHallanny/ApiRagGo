package mcp

import (
	"Api_Rag/models"
	"Api_Rag/services"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerTools() {
	// Tool 1: Busca Semântica RAG no catálogo
	searchTool := mcp.NewTool(
		"search_albums_semantic",
		mcp.WithDescription("Realiza busca semântica em linguagem natural no catálogo de discos com pgvector e HNSW. Retorna álbuns semanticamente mais próximos da descrição informada com score de similaridade."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Termo de busca, tema, instrumentos, clima ou descrição em linguagem natural")),
		mcp.WithInteger("limit", mcp.Description("Quantidade máxima de álbuns a retornar (padrão: 5)")),
		mcp.WithNumber("min_similarity", mcp.Description("Limiar mínimo de similaridade por cosseno entre 0.0 e 1.0 (padrão: 0.60)")),
		mcp.WithReadOnlyHintAnnotation(true),
	)

	s.mcpServer.AddTool(searchTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("O parâmetro 'query' é obrigatório"), nil
		}

		limit := req.GetInt("limit", 5)
		if limit <= 0 {
			limit = 5
		}

		minSim := req.GetFloat("min_similarity", 0.60)
		if minSim < 0.0 || minSim > 1.0 {
			minSim = 0.60
		}

		results, err := s.albumService.SearchSimilar(ctx, query, limit, minSim)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Falha na busca semântica: %v", err)), nil
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("Nenhum álbum encontrado com os critérios de similaridade informados."), nil
		}

		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao serializar resultado: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Tool 2: Obter Álbum por ID
	getAlbumTool := mcp.NewTool(
		"get_album_by_id",
		mcp.WithDescription("Busca os detalhes completos de um álbum específico através do seu ID único."),
		mcp.WithString("id", mcp.Required(), mcp.Description("ID único do álbum")),
		mcp.WithReadOnlyHintAnnotation(true),
	)

	s.mcpServer.AddTool(getAlbumTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("O parâmetro 'id' é obrigatório"), nil
		}

		album, err := s.albumService.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, services.ErrNotFound) {
				return mcp.NewToolResultError(fmt.Sprintf("Álbum com ID '%s' não foi encontrado.", id)), nil
			}
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao consultar álbum: %v", err)), nil
		}

		jsonData, err := json.MarshalIndent(album, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao serializar dados do álbum: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Tool 3: Listar Todos os Álbuns do Catálogo
	listAlbumsTool := mcp.NewTool(
		"list_albums",
		mcp.WithDescription("Lista todos os álbuns de música cadastrados no catálogo do repositório."),
		mcp.WithReadOnlyHintAnnotation(true),
	)

	s.mcpServer.AddTool(listAlbumsTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		albums, err := s.albumService.GetAll(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao listar álbuns: %v", err)), nil
		}

		if len(albums) == 0 {
			return mcp.NewToolResultText("Nenhum álbum cadastrado no catálogo atualmente."), nil
		}

		jsonData, err := json.MarshalIndent(albums, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao serializar álbuns: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	})

	// Tool 4: Criar/Ingerir Novo Álbum com Embeddings Automáticos
	createAlbumTool := mcp.NewTool(
		"create_album",
		mcp.WithDescription("Cadastra um novo álbum no catálogo. Gera automaticamente o embedding vetorial de 768 dimensões com Google Gemini (text-embedding-004) para habilitar busca semântica imediata."),
		mcp.WithString("id", mcp.Description("ID único opcional para o álbum")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Título do álbum")),
		mcp.WithString("artist", mcp.Required(), mcp.Description("Nome do artista ou banda")),
		mcp.WithNumber("price", mcp.Required(), mcp.Description("Preço de venda do álbum")),
		mcp.WithString("description", mcp.Description("Descrição rica do estilo musical, histórico, instrumentos e faixas")),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
	)

	s.mcpServer.AddTool(createAlbumTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("O parâmetro 'title' é obrigatório"), nil
		}

		artist, err := req.RequireString("artist")
		if err != nil {
			return mcp.NewToolResultError("O parâmetro 'artist' é obrigatório"), nil
		}

		price, err := req.RequireFloat("price")
		if err != nil {
			return mcp.NewToolResultError("O parâmetro 'price' é obrigatório"), nil
		}

		id := req.GetString("id", "")
		desc := req.GetString("description", "")

		newAlbum := models.Album{
			ID:          id,
			Title:       title,
			Artist:      artist,
			Price:       price,
			Description: desc,
		}

		created, err := s.albumService.Create(ctx, newAlbum)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Falha ao criar álbum: %v", err)), nil
		}

		jsonData, err := json.MarshalIndent(created, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Erro ao serializar álbum criado: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Álbum cadastrado com sucesso e indexado no pgvector:\n%s", string(jsonData))), nil
	})
}
