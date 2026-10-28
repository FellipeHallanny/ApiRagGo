package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pgvector/pgvector-go"
)

type EmbeddingService struct {
	apiKey    string
	client    *http.Client
	modelName string
}

type listModelsResponse struct {
	Models []struct {
		Name                       string   `json:"name"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
}

type geminiEmbedRequest struct {
	Content              geminiContent `json:"content"`
	OutputDimensionality *int          `json:"outputDimensionality,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// NewEmbeddingService instancia o serviço Gemini descobrindo automaticamente o modelo de embedding disponível
func NewEmbeddingService() (*EmbeddingService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("variável de ambiente GEMINI_API_KEY não definida")
	}

	svc := &EmbeddingService{
		apiKey: apiKey,
		client: &http.Client{},
	}

	// Descobre o modelo de embedding suportado pela chave
	model, err := svc.discoverEmbeddingModel()
	if err != nil {
		log.Printf("Aviso: falha na descoberta automática de modelos (%v). Usando 'text-embedding-004' como padrão.", err)
		svc.modelName = "models/text-embedding-004"
	} else {
		svc.modelName = model
		log.Printf("Modelo de embedding selecionado: %s", svc.modelName)
	}

	return svc, nil
}

func (s *EmbeddingService) discoverEmbeddingModel() (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", s.apiKey)
	resp, err := s.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var list listModelsResponse
	if err := json.Unmarshal(body, &list); err != nil {
		return "", err
	}

	var candidates []string
	for _, m := range list.Models {
		for _, method := range m.SupportedGenerationMethods {
			if method == "embedContent" {
				candidates = append(candidates, m.Name)
				// Preferência por text-embedding-004 se estiver na lista
				if strings.Contains(m.Name, "text-embedding-004") {
					return m.Name, nil
				}
			}
		}
	}

	if len(candidates) > 0 {
		return candidates[0], nil
	}

	return "", fmt.Errorf("nenhum modelo suporta 'embedContent'")
}

func (s *EmbeddingService) Close() {
}

// GenerateEmbedding envia o texto para a API REST do Gemini fixando a dimensão em 768
func (s *EmbeddingService) GenerateEmbedding(ctx context.Context, text string) (pgvector.Vector, error) {
	modelPath := s.modelName
	if modelPath == "" {
		modelPath = "models/text-embedding-004"
	}

	modelClean := strings.TrimPrefix(modelPath, "models/")
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", modelClean, s.apiKey)

	dim768 := 768
	reqPayload := geminiEmbedRequest{
		Content: geminiContent{
			Parts: []geminiPart{
				{Text: text},
			},
		},
		OutputDimensionality: &dim768,
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("falha ao serializar payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("falha ao criar requisição HTTP: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("falha na chamada HTTP para o Gemini: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return pgvector.Vector{}, fmt.Errorf("falha ao ler resposta do Gemini: %w", err)
	}

	// Se o modelo não aceitar o parâmetro outputDimensionality (ex: modelos legados), tenta sem ele
	if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(respBytes), "outputDimensionality") {
		reqPayload.OutputDimensionality = nil
		bodyBytes, _ = json.Marshal(reqPayload)
		reqRetry, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
		reqRetry.Header.Set("Content-Type", "application/json")
		respRetry, errRetry := s.client.Do(reqRetry)
		if errRetry == nil {
			defer respRetry.Body.Close()
			resp = respRetry
			respBytes, _ = io.ReadAll(respRetry.Body)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return pgvector.Vector{}, fmt.Errorf("Gemini retornou status %d: %s", resp.StatusCode, string(respBytes))
	}

	var geminiResp geminiEmbedResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return pgvector.Vector{}, fmt.Errorf("falha ao deserializar resposta do Gemini: %w", err)
	}

	if geminiResp.Error != nil {
		return pgvector.Vector{}, fmt.Errorf("erro da API Gemini: %s", geminiResp.Error.Message)
	}

	vals := geminiResp.Embedding.Values
	if len(vals) == 0 {
		return pgvector.Vector{}, fmt.Errorf("Gemini retornou vetor vazio")
	}

	// Trunca para 768 caso o modelo retorne 3072 e não suporte o parâmetro outputDimensionality
	if len(vals) > 768 {
		vals = vals[:768]
	}

	return pgvector.NewVector(vals), nil
}
