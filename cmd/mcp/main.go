package main

import (
	"Api_Rag/db"
	"Api_Rag/services"
	"Api_Rag/services/gemini"
	"Api_Rag/store"
	transportmcp "Api_Rag/transport/mcp"
	"log"
	"os"
)

func main() {
	// CRÍTICO NO PROTOCOLO MCP VIA STDIO:
	// O stdout é o canal exclusivo de comunicação JSON-RPC com o cliente MCP (Claude Desktop, Cursor, etc.).
	// Todos os logs operacionais, advertências e diagnósticos DEVEM ir para o stderr.
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	log.Println("[ApiRag MCP] Inicializando servidor em modo STDIO...")

	// 1. Conexão com o PostgreSQL + pgvector
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("[ApiRag MCP] Erro crítico ao conectar no banco de dados: %v", err)
	}
	defer func() {
		log.Println("[ApiRag MCP] Fechando conexão com PostgreSQL...")
		_ = database.Close()
	}()
	log.Println("[ApiRag MCP] Conexão com PostgreSQL (pgvector) ativa.")

	// 2. Inicialização do serviço de embeddings Gemini (text-embedding-004)
	geminiService, err := gemini.NewEmbeddingService()
	if err != nil {
		log.Printf("[ApiRag MCP] AVISO: Gemini Embedding Service desabilitado ou sem chave (%v).", err)
		log.Println("[ApiRag MCP] A busca semântica e geração automática de vetores ficarão indisponíveis até definir GEMINI_API_KEY.")
	} else {
		defer func() {
			log.Println("[ApiRag MCP] Encerrando cliente Gemini...")
			geminiService.Close()
		}()
		log.Println("[ApiRag MCP] Serviço Gemini Embeddings inicializado com sucesso.")
	}

	// 3. Camada de Domínio / Regras de Negócio
	albumStore := store.NewAlbumStore(database)
	albumService := services.NewAlbumService(albumStore, geminiService)

	// 4. Inicialização do Servidor MCP
	mcpServer := transportmcp.NewServer(albumService)

	// 5. Escuta no Stdio (bloqueia até encerramento do processo pelo cliente MCP)
	log.Println("[ApiRag MCP] Servidor MCP pronto e ouvindo comandos JSON-RPC via STDIN...")
	if err := mcpServer.ServeStdio(); err != nil {
		log.Fatalf("[ApiRag MCP] Erro fatal durante execução do servidor MCP: %v", err)
	}
}
