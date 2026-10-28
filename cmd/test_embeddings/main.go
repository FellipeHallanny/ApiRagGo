package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type AlbumPayload struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Embedding   []float32 `json:"embedding"`
}

type SearchPayload struct {
	Embedding []float32 `json:"embedding"`
	Limit     int       `json:"limit"`
}

func generateEmbedding(baseValue float32) []float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	emb := make([]float32, 768)
	for i := range emb {
		// Gera valores ao redor de baseValue para testar proximidade semântica
		emb[i] = baseValue + (r.Float32() * 0.05)
	}
	return emb
}

func main() {
	baseURL := "http://localhost:8080"

	// 1. Criar vetores sintéticos para teste
	// Jazz coltrane e Miles Davis terão vetores próximos (ao redor de 0.8)
	// Rock Pink Floyd terá vetor bem distante (ao redor de 0.1)
	jazzColtraneEmb := generateEmbedding(0.80)
	jazzMilesEmb := generateEmbedding(0.78)
	rockFloydEmb := generateEmbedding(0.10)

	albums := []AlbumPayload{
		{
			ID:          "1",
			Title:       "Blue Train",
			Artist:      "John Coltrane",
			Price:       56.99,
			Description: "Álbum seminal de hard bop jazz lançado pela Blue Note em 1957.",
			Embedding:   jazzColtraneEmb,
		},
		{
			ID:          "2",
			Title:       "Kind of Blue",
			Artist:      "Miles Davis",
			Price:       62.50,
			Description: "Obra-prima do jazz modal de 1959.",
			Embedding:   jazzMilesEmb,
		},
		{
			ID:          "3",
			Title:       "The Dark Side of the Moon",
			Artist:      "Pink Floyd",
			Price:       70.00,
			Description: "Álbum clássico de rock progressivo e psicodélico de 1973.",
			Embedding:   rockFloydEmb,
		},
	}

	fmt.Println("--- 1. Cadastrando Álbuns com Embedding (768 dimensões) ---")
	for _, album := range albums {
		body, _ := json.Marshal(album)
		resp, err := http.Post(baseURL+"/albums", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Printf("Erro ao conectar à API (%s). Certifique-se de que a API está rodando.\n", err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Cadastrado '%s': Status %d\n", album.Title, resp.StatusCode)
	}

	fmt.Println("\n--- 2. Testando Busca Semântica por Proximidade de Jazz ---")
	// Vetor de busca simulando uma busca por Jazz (ao redor de 0.79)
	queryEmb := generateEmbedding(0.79)
	searchBody, _ := json.Marshal(SearchPayload{
		Embedding: queryEmb,
		Limit:     3,
	})

	resp, err := http.Post(baseURL+"/albums/search", "application/json", bytes.NewBuffer(searchBody))
	if err != nil {
		fmt.Printf("Erro na busca: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)
	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, respData, "", "  ")

	fmt.Printf("Resultado da busca semântica (ordenado por maior similaridade):\n%s\n", prettyJSON.String())
}
