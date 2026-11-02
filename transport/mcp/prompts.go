package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerPrompts() {
	// Prompt: Curadoria musical inteligente usando RAG
	curatorPrompt := mcp.NewPrompt(
		"curate_music_by_mood",
		mcp.WithPromptTitle("Curador Musical Inteligente RAG"),
		mcp.WithPromptDescription("Orienta o modelo de IA a atuar como um curador musical refinado, realizando busca semântica na base e elaborando uma recomendação com justificativas contextuais."),
		mcp.WithArgument("mood", mcp.RequiredArgument(), mcp.ArgumentDescription("Clima, sentimento ou atividade pretendida (ex: 'foco profundo para programar', 'noite chuvosa relaxante')")),
		mcp.WithArgument("genre", mcp.ArgumentDescription("Gênero musical preferencial opcional (ex: 'jazz', 'rock', 'blues')")),
	)

	s.mcpServer.AddPrompt(curatorPrompt, func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		mood := req.Params.Arguments["mood"]
		genre := req.Params.Arguments["genre"]

		instruction := fmt.Sprintf(
			"Você é um curador musical especialista. O usuário está buscando recomendações para o seguinte momento:\n"+
				"- Clima/Atividade: %s\n",
			mood,
		)

		if genre != "" {
			instruction += fmt.Sprintf("- Preferência de Gênero: %s\n", genre)
		}

		instruction += "\nSiga rigorosamente estas etapas:\n" +
			"1. Utilize a ferramenta 'search_albums_semantic' com termos representativos do clima e instrumentos para encontrar os melhores álbuns em nosso catálogo pgvector.\n" +
			"2. Analise os álbuns retornados e o score de similaridade.\n" +
			"3. Apresente uma seleção justificada explicando por que a sonoridade, instrumentos ou contexto histórico do disco combinam com o momento solicitado.\n" +
			"4. Indique o preço e metadados de cada álbum recomendado."

		return mcp.NewGetPromptResult(
			"Instruções para curadoria musical baseada em RAG semântico",
			[]mcp.PromptMessage{
				mcp.NewPromptMessage(
					mcp.RoleUser,
					mcp.NewTextContent(instruction),
				),
			},
		), nil
	})
}
