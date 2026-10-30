package grpc_test

import (
	ragv1 "Api_Rag/gen/rag/v1"
	"Api_Rag/models"
	"Api_Rag/services"
	"Api_Rag/store"
	transportgrpc "Api_Rag/transport/grpc"
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCServer_ReflectionAndContract(t *testing.T) {
	// Cria listener em memória/porta efêmera
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("falha ao abrir porta para teste: %v", err)
	}
	defer lis.Close()

	// Inicializa mock/instância de serviço sem DB real para validar wiring
	albumService := services.NewAlbumService((*store.AlbumStore)(nil), nil)
	srv := transportgrpc.NewGRPCServer(albumService)

	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	// Conecta como cliente gRPC
	conn, err := grpc.NewClient(
		lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("falha ao conectar no gRPC server de teste: %v", err)
	}
	defer conn.Close()

	client := ragv1.NewAlbumServiceClient(conn)

	// Validação de argumento inválido no CreateAlbum
	_, err = client.CreateAlbum(context.Background(), &ragv1.CreateAlbumRequest{
		Title: "",
	})
	if err == nil {
		t.Fatalf("esperava erro de validação para título vazio, mas obteve nil")
	}

	// Validação de busca vazia
	_, err = client.SearchSimilar(context.Background(), &ragv1.SearchSimilarRequest{
		Query: "",
	})
	if err == nil {
		t.Fatalf("esperava erro para query vazia, mas obteve nil")
	}
}

func TestModelToProto(t *testing.T) {
	album := &models.Album{
		ID:          "123",
		Title:       "Kind of Blue",
		Artist:      "Miles Davis",
		Price:       59.99,
		Description: "Jazz clássico",
	}

	protoAlbum := transportgrpc.ModelToProto(album)
	if protoAlbum == nil {
		t.Fatalf("esperava protoAlbum preenchido, obteve nil")
	}
	if protoAlbum.GetTitle() != "Kind of Blue" {
		t.Fatalf("título inconsistente: esperado 'Kind of Blue', obtido '%s'", protoAlbum.GetTitle())
	}
}
