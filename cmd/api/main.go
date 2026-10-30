package main

import (
	"Api_Rag/app"
	"Api_Rag/db"
	"Api_Rag/services"
	"Api_Rag/services/gemini"
	"Api_Rag/store"
	transportgrpc "Api_Rag/transport/grpc"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Conexão com PostgreSQL (pgvector)
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar no banco de dados: %v", err)
	}
	defer database.Close()
	log.Println("Conexão com PostgreSQL (pgvector) estabelecida!")

	// 2. Inicialização do serviço Gemini
	geminiService, err := gemini.NewEmbeddingService()
	if err != nil {
		log.Printf("AVISO: Gemini Embedding Service não inicializado (%v). Defina GEMINI_API_KEY.", err)
	} else {
		defer geminiService.Close()
		log.Println("Serviço Gemini (text-embedding-004, 768 dimensões) inicializado!")
	}

	// 3. Camada de Domínio / Regras de Negócio
	albumStore := store.NewAlbumStore(database)
	albumService := services.NewAlbumService(albumStore, geminiService)

	// 4. Inicialização do Servidor gRPC
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Erro ao abrir porta gRPC :%s: %v", grpcPort, err)
	}

	grpcServer := transportgrpc.NewGRPCServer(albumService)
	go func() {
		log.Printf("🚀 Servidor gRPC escutando em :%s (Reflection e Health ativos)", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Erro ao executar servidor gRPC: %v", err)
		}
	}()

	// 5. Inicialização da API HTTP/REST (Gin) para compatibilidade dual-stack
	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	ginRouter := app.NewApp(database, geminiService)
	httpServer := &http.Server{
		Addr:    ":" + httpPort,
		Handler: ginRouter,
	}

	go func() {
		log.Printf("🌐 Servidor HTTP REST (Gin) escutando em :%s", httpPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Erro ao executar servidor HTTP: %v", err)
		}
	}()

	// 6. Graceful Shutdown cooperativo
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Encerrando servidores graciosamente...")

	// Encerra gRPC
	grpcServer.GracefulStop()
	log.Println("Servidor gRPC encerrado.")

	// Encerra HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Erro ao encerrar servidor HTTP: %v", err)
	}
	log.Println("Servidor HTTP encerrado com sucesso.")
}
