package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPTester struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	reqID  int
}

func NewMCPTester(binaryPath string) (*MCPTester, error) {
	cmd := exec.Command(binaryPath)
	cmd.Stderr = os.Stderr // Redireciona logs operacionais do MCP para o Stderr do console

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("falha ao iniciar processo MCP: %w", err)
	}

	return &MCPTester{
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
		reqID:  1,
	}, nil
}

func (t *MCPTester) Call(method string, params map[string]any) (*RPCResponse, error) {
	req := RPCRequest{
		JSONRPC: "2.0",
		ID:      t.reqID,
		Method:  method,
		Params:  params,
	}
	t.reqID++

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição RPC: %w", err)
	}

	line, err := t.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta RPC: %w", err)
	}

	var resp RPCResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &resp); err != nil {
		return nil, fmt.Errorf("resposta JSON inválida (%s): %w", line, err)
	}

	return &resp, nil
}

func (t *MCPTester) Close() {
	_ = t.stdin.Close()
	_ = t.cmd.Process.Kill()
}

func main() {
	fmt.Println("==================================================================")
	fmt.Println("     TESTE LOCAL DO SERVIDOR MCP (STDIO) - Api_Rag")
	fmt.Println("==================================================================")

	binaryPath := filepath.Join("bin", "mcp.exe")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		fmt.Println("[1/5] Compilando binário bin/mcp.exe...")
		buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/mcp")
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			log.Fatalf("Erro ao compilar MCP: %v", err)
		}
	}

	fmt.Printf("Iniciando processo MCP (%s)...\n\n", binaryPath)
	client, err := NewMCPTester(binaryPath)
	if err != nil {
		log.Fatalf("Erro ao iniciar testador: %v", err)
	}
	defer client.Close()

	// 1. Handshake Inicial (initialize)
	fmt.Println("▶ [PASSO 1] Enviando 'initialize' (Handshake do Protocolo MCP)...")
	initResp, err := client.Call("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "mcp-local-tester",
			"version": "1.0.0",
		},
	})
	if err != nil {
		log.Fatalf("Falha no handshake: %v", err)
	}
	printJSON("Resultado do Handshake", initResp.Result)

	// 2. Listagem de Tools (tools/list)
	fmt.Println("\n▶ [PASSO 2] Consultando lista de ferramentas ('tools/list')...")
	toolsResp, err := client.Call("tools/list", map[string]any{})
	if err != nil {
		log.Fatalf("Falha ao listar ferramentas: %v", err)
	}
	printJSON("Ferramentas Registradas", toolsResp.Result)

	// 3. Execução de Tool: list_albums (tools/call)
	fmt.Println("\n▶ [PASSO 3] Executando Tool 'list_albums'...")
	listResp, err := client.Call("tools/call", map[string]any{
		"name":      "list_albums",
		"arguments": map[string]any{},
	})
	if err != nil {
		log.Fatalf("Falha ao chamar list_albums: %v", err)
	}
	printJSON("Retorno da Tool 'list_albums'", listResp.Result)

	// 4. Leitura de Resource: albums://catalog (resources/read)
	fmt.Println("\n▶ [PASSO 4] Lendo Recurso 'albums://catalog' ('resources/read')...")
	resResp, err := client.Call("resources/read", map[string]any{
		"uri": "albums://catalog",
	})
	if err != nil {
		log.Fatalf("Falha ao ler recurso: %v", err)
	}
	printJSON("Conteúdo do Recurso 'albums://catalog'", resResp.Result)

	// 5. Consulta de Prompt: curate_music_by_mood (prompts/get)
	fmt.Println("\n▶ [PASSO 5] Obtendo Prompt 'curate_music_by_mood' ('prompts/get')...")
	promptResp, err := client.Call("prompts/get", map[string]any{
		"name": "curate_music_by_mood",
		"arguments": map[string]string{
			"mood":  "foco profundo para codificar em Go",
			"genre": "jazz instrumental acústico",
		},
	})
	if err != nil {
		log.Fatalf("Falha ao obter prompt: %v", err)
	}
	printJSON("Template do Prompt de Curadoria", promptResp.Result)

	time.Sleep(200 * time.Millisecond)
	fmt.Println("\n==================================================================")
	fmt.Println("  TODOS OS TESTES LOCAIS DO MCP EXECUTADOS COM SUCESSO!")
	fmt.Println("==================================================================")
}

func printJSON(label string, raw json.RawMessage) {
	var pretty any
	if err := json.Unmarshal(raw, &pretty); err == nil {
		formatted, _ := json.MarshalIndent(pretty, "   ", "  ")
		fmt.Printf("   [%s]:\n   %s\n", label, string(formatted))
	} else {
		fmt.Printf("   [%s]: %s\n", label, string(raw))
	}
}
