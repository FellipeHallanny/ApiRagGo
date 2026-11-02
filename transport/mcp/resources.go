package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerResources() {
	// Resource 1: Catálogo geral de álbuns
	catalogResource := mcp.NewResource(
		"albums://catalog",
		"Catálogo Completo de Álbuns",
		mcp.WithResourceDescription("Arquivo em formato JSON contendo todos os álbuns cadastrados no repositório."),
		mcp.WithMIMEType("application/json"),
	)

	s.mcpServer.AddResource(catalogResource, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		albums, err := s.albumService.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar catálogo para recurso: %w", err)
		}

		jsonData, err := json.MarshalIndent(albums, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar recurso de catálogo: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	})

	// Resource Template 2: Álbum individual por ID dinâmico (albums://{id})
	albumTemplate := mcp.NewResourceTemplate(
		"albums://{id}",
		"Detalhes do Álbum por ID",
		mcp.WithTemplateDescription("Acessa diretamente as informações e metadados de um álbum musical pelo seu ID."),
		mcp.WithTemplateMIMEType("application/json"),
	)

	s.mcpServer.AddResourceTemplate(albumTemplate, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
		id := strings.TrimPrefix(req.Params.URI, "albums://")
		if id == "" || id == "catalog" {
			return nil, fmt.Errorf("URI de álbum inválida: %s", req.Params.URI)
		}

		album, err := s.albumService.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("álbum não encontrado para o recurso (%s): %w", id, err)
		}

		jsonData, err := json.MarshalIndent(album, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar recurso de álbum: %w", err)
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(jsonData),
			},
		}, nil
	})
}
