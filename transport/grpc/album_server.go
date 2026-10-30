package grpc

import (
	ragv1 "Api_Rag/gen/rag/v1"
	"Api_Rag/models"
	"Api_Rag/services"
	"context"
	"errors"

	"github.com/pgvector/pgvector-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AlbumGrpcServer struct {
	ragv1.UnimplementedAlbumServiceServer
	service *services.AlbumService
}

func NewAlbumGrpcServer(service *services.AlbumService) *AlbumGrpcServer {
	return &AlbumGrpcServer{
		service: service,
	}
}

// ModelToProto converte modelo de domínio para proto
func ModelToProto(a *models.Album) *ragv1.Album {
	if a == nil {
		return nil
	}
	return &ragv1.Album{
		Id:          a.ID,
		Title:       a.Title,
		Artist:      a.Artist,
		Price:       a.Price,
		Description: a.Description,
		Embedding:   a.Embedding.Slice(),
	}
}

func (s *AlbumGrpcServer) GetAlbums(ctx context.Context, req *ragv1.GetAlbumsRequest) (*ragv1.GetAlbumsResponse, error) {
	albums, err := s.service.GetAll(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar álbuns: %v", err)
	}

	protoAlbums := make([]*ragv1.Album, 0, len(albums))
	for i := range albums {
		protoAlbums = append(protoAlbums, ModelToProto(&albums[i]))
	}

	return &ragv1.GetAlbumsResponse{
		Albums: protoAlbums,
	}, nil
}

func (s *AlbumGrpcServer) GetAlbumByID(ctx context.Context, req *ragv1.GetAlbumByIDRequest) (*ragv1.GetAlbumByIDResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "o ID do álbum é obrigatório")
	}

	album, err := s.service.GetByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "álbum com ID '%s' não encontrado", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "falha ao buscar álbum: %v", err)
	}

	return &ragv1.GetAlbumByIDResponse{
		Album: ModelToProto(album),
	}, nil
}

func (s *AlbumGrpcServer) CreateAlbum(ctx context.Context, req *ragv1.CreateAlbumRequest) (*ragv1.CreateAlbumResponse, error) {
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "o título do álbum é obrigatório")
	}
	if req.GetArtist() == "" {
		return nil, status.Error(codes.InvalidArgument, "o artista do álbum é obrigatório")
	}
	if req.GetPrice() < 0 {
		return nil, status.Error(codes.InvalidArgument, "o preço não pode ser negativo")
	}

	domainAlbum := models.Album{
		Title:       req.GetTitle(),
		Artist:      req.GetArtist(),
		Price:       req.GetPrice(),
		Description: req.GetDescription(),
	}

	if len(req.GetEmbedding()) > 0 {
		domainAlbum.Embedding = pgvector.NewVector(req.GetEmbedding())
	}

	created, err := s.service.Create(ctx, domainAlbum)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao criar álbum: %v", err)
	}

	return &ragv1.CreateAlbumResponse{
		Album: ModelToProto(created),
	}, nil
}

func (s *AlbumGrpcServer) SearchSimilar(ctx context.Context, req *ragv1.SearchSimilarRequest) (*ragv1.SearchSimilarResponse, error) {
	if req.GetQuery() == "" {
		return nil, status.Error(codes.InvalidArgument, "a consulta de busca não pode ser vazia")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 5
	}

	results, err := s.service.SearchSimilar(ctx, req.GetQuery(), limit, req.GetMinSimilarity())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha na busca por similaridade RAG: %v", err)
	}

	protoResults := make([]*ragv1.AlbumSearchResult, 0, len(results))
	for _, res := range results {
		protoResults = append(protoResults, &ragv1.AlbumSearchResult{
			Album:      ModelToProto(&res.Album),
			Similarity: res.Similarity,
		})
	}

	return &ragv1.SearchSimilarResponse{
		Results: protoResults,
	}, nil
}

func (s *AlbumGrpcServer) SearchSimilarStream(req *ragv1.SearchSimilarRequest, stream ragv1.AlbumService_SearchSimilarStreamServer) error {
	if req.GetQuery() == "" {
		return status.Error(codes.InvalidArgument, "a consulta de busca não pode ser vazia")
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 5
	}

	results, err := s.service.SearchSimilar(stream.Context(), req.GetQuery(), limit, req.GetMinSimilarity())
	if err != nil {
		return status.Errorf(codes.Internal, "falha na busca por similaridade RAG: %v", err)
	}

	for _, res := range results {
		chunk := &ragv1.SearchSimilarChunk{
			Result: &ragv1.AlbumSearchResult{
				Album:      ModelToProto(&res.Album),
				Similarity: res.Similarity,
			},
		}
		if err := stream.Send(chunk); err != nil {
			return status.Errorf(codes.Internal, "falha ao enviar stream de dados: %v", err)
		}
	}

	return nil
}
