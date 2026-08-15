package estoque

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type IAClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewIAClient() *IAClient {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-flash-latest"
	}
	return &IAClient{
		apiKey:     os.Getenv("GEMINI_API_KEY"),
		model:      model,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type sugestaoIA struct {
	Descricao string   `json:"descricao"`
	Similares []string `json:"similares"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
}

func (c *IAClient) Sugerir(ctx context.Context, codigo string, catalogo []Produto) (string, []string, error) {
	if c.apiKey == "" {
		return "", nil, ErrIAIndisponivel
	}

	var catalogoTexto strings.Builder
	for _, p := range catalogo {
		fmt.Fprintf(&catalogoTexto, "- %s: %s\n", p.Codigo, p.Descricao)
	}
	if catalogoTexto.Len() == 0 {
		catalogoTexto.WriteString("(catalogo vazio)\n")
	}

	prompt := fmt.Sprintf(`Voce ajuda a cadastrar produtos num sistema de estoque.

Codigo do novo produto: %s

Catalogo de produtos ja cadastrados (codigo: descricao):
%s
Tarefa:
1. Sugira uma descricao curta e realista para o produto de codigo %q, baseada no padrao dos codigos/descricoes existentes.
2. Aponte, no maximo 3, codigos do catalogo acima que sejam produtos similares (mesma categoria/familia). Se nao houver nenhum similar, devolva lista vazia.

Responda APENAS com JSON valido, sem markdown, no formato exato:
{"descricao": "...", "similares": ["COD1", "COD2"]}`, codigo, catalogoTexto.String(), codigo)

	reqBody, err := json.Marshal(geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
	})
	if err != nil {
		return "", nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, ErrIAIndisponivel
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, ErrIAIndisponivel
	}
	if resp.StatusCode != http.StatusOK {
		return "", nil, ErrIAIndisponivel
	}

	var gemResp geminiResponse
	if err := json.Unmarshal(corpo, &gemResp); err != nil || len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return "", nil, ErrIAIndisponivel
	}

	texto := strings.TrimSpace(gemResp.Candidates[0].Content.Parts[0].Text)
	texto = strings.TrimPrefix(texto, "```json")
	texto = strings.TrimPrefix(texto, "```")
	texto = strings.TrimSuffix(texto, "```")
	texto = strings.TrimSpace(texto)

	var sugestao sugestaoIA
	if err := json.Unmarshal([]byte(texto), &sugestao); err != nil {
		return "", nil, ErrIAIndisponivel
	}

	return sugestao.Descricao, sugestao.Similares, nil
}
