package main

import (
	"Api_Rag/app"
	"Api_Rag/db"
	"Api_Rag/services/gemini"
	"log"
)

func main() {
	// 1. Conexão com PostgreSQL
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de dados: %v", err)
	}
	defer database.Close()
	log.Println("Conexão com PostgreSQL (pgvector) estabelecida!")

	// 2. Inicialização do serviço Gemini
	geminiService, err := gemini.NewEmbeddingService()
	if err != nil {
		log.Printf("AVISO: Gemini Embedding Service não inicializado (%v). Certifique-se de definir GEMINI_API_KEY.", err)
	} else {
		defer geminiService.Close()
		log.Println("Serviço Gemini (text-embedding-004, 768 dimensões) inicializado!")
	}

	// 3. Inicialização da API
	router := app.NewApp(database, geminiService)
	if err := router.Run("localhost:8080"); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
