package mcp_test

import (
	"Api_Rag/services"
	"Api_Rag/store"
	transportmcp "Api_Rag/transport/mcp"
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMCPServer_RegistrationAndTools(t *testing.T) {
	// Cria o serviço sem conexão com banco real para testes de wiring do protocolo
	albumService := services.NewAlbumService((*store.AlbumStore)(nil), nil)
	srv := transportmcp.NewServer(albumService)
	mcpSrv := srv.MCPServer()

	// 1. Testa presença das tools esperadas
	expectedTools := []string{
		"search_albums_semantic",
		"get_album_by_id",
		"list_albums",
		"create_album",
	}

	for _, toolName := range expectedTools {
		tool := mcpSrv.GetTool(toolName)
		if tool == nil {
			t.Fatalf("Esperava tool '%s' registrada no MCP Server, mas não foi encontrada", toolName)
		}
	}

	// 2. Testa validação de parâmetros obrigatórios da tool search_albums_semantic
	ctx := context.Background()
	searchTool := mcpSrv.GetTool("search_albums_semantic")

	// Chamada sem 'query'
	reqWithoutQuery := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "search_albums_semantic",
			Arguments: map[string]any{},
		},
	}
	res, err := searchTool.Handler(ctx, reqWithoutQuery)
	if err != nil {
		t.Fatalf("Handler da tool retornou erro inesperado: %v", err)
	}
	if !res.IsError {
		t.Fatalf("Esperava resultado de erro para 'search_albums_semantic' sem query, mas obteve sucesso")
	}

	// 3. Testa validação de create_album sem parâmetros obrigatórios
	createTool := mcpSrv.GetTool("create_album")
	reqWithoutTitle := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_album",
			Arguments: map[string]any{
				"artist": "John Coltrane",
				"price":  45.0,
			},
		},
	}
	res, err = createTool.Handler(ctx, reqWithoutTitle)
	if err != nil {
		t.Fatalf("Handler retornou erro inesperado: %v", err)
	}
	if !res.IsError {
		t.Fatalf("Esperava erro ao criar álbum sem título")
	}

	// 4. Testa validação de get_album_by_id sem 'id'
	getTool := mcpSrv.GetTool("get_album_by_id")
	reqWithoutID := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "get_album_by_id",
			Arguments: map[string]any{},
		},
	}
	res, err = getTool.Handler(ctx, reqWithoutID)
	if err != nil {
		t.Fatalf("Handler retornou erro inesperado: %v", err)
	}
	if !res.IsError {
		t.Fatalf("Esperava erro ao buscar álbum sem ID")
	}
}

func TestMCPServer_Prompts(t *testing.T) {
	albumService := services.NewAlbumService((*store.AlbumStore)(nil), nil)
	srv := transportmcp.NewServer(albumService)
	mcpSrv := srv.MCPServer()

	prompts := mcpSrv.ListPrompts()
	prompt, exists := prompts["curate_music_by_mood"]
	if !exists {
		t.Fatalf("Prompt 'curate_music_by_mood' não encontrado no servidor MCP")
	}

	req := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "curate_music_by_mood",
			Arguments: map[string]string{
				"mood":  "foco profundo para programação",
				"genre": "jazz instrumental",
			},
		},
	}

	result, err := prompt.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Erro ao executar handler do prompt: %v", err)
	}

	if len(result.Messages) == 0 {
		t.Fatalf("Esperava mensagens no resultado do prompt, obteve 0")
	}

	text := result.Messages[0].Content.(mcp.TextContent).Text
	if !strings.Contains(text, "foco profundo para programação") {
		t.Errorf("Prompt não incluiu o argumento de mood esperado. Texto: %s", text)
	}
	if !strings.Contains(text, "jazz instrumental") {
		t.Errorf("Prompt não incluiu o gênero musical esperado. Texto: %s", text)
	}
}
