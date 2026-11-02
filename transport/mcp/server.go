package mcp

import (
	"Api_Rag/services"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

// Server encapsula o servidor MCP e suas dependências de negócio.
type Server struct {
	mcpServer    *server.MCPServer
	albumService *services.AlbumService
}

// NewServer cria e configura uma nova instância do servidor MCP para o Api_Rag.
func NewServer(albumService *services.AlbumService) *Server {
	mcpSrv := server.NewMCPServer(
		"api-rag-mcp",
		"1.0.0",
		server.WithTitle("ApiRag Music Catalog & Semantic Search MCP Server"),
		server.WithDescription("Servidor MCP para busca semântica RAG e gerenciamento do catálogo de álbuns musicais com PostgreSQL pgvector e Google Gemini."),
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithRecovery(),
	)

	s := &Server{
		mcpServer:    mcpSrv,
		albumService: albumService,
	}

	// Registra os componentes MCP
	s.registerTools()
	s.registerResources()
	s.registerPrompts()

	return s
}

// MCPServer retorna a instância subjacente do servidor MCP (útil para testes unitários).
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// ServeStdio inicializa a escuta no canal Standard Input/Output (JSON-RPC) para clientes MCP locais.
// Importante: garante que logs gerais sejam direcionados para stderr para não poluir o stdout.
func (s *Server) ServeStdio() error {
	log.SetOutput(os.Stderr)
	log.Println("[MCP] Inicializando ApiRag MCP Server no modo STDIO...")

	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("falha na execução do servidor MCP em modo STDIO: %w", err)
	}

	log.Println("[MCP] Servidor MCP STDIO finalizado com sucesso.")
	return nil
}
