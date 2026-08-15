package faturamento

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"korp/internal/httpx"
)

type ItemBaixaPayload struct {
	ProdutoID  int64 `json:"produto_id"`
	Quantidade int   `json:"quantidade"`
}

type baixarRequestPayload struct {
	Itens []ItemBaixaPayload `json:"itens"`
}

type EstoqueClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewEstoqueClient(baseURL string) *EstoqueClient {
	return &EstoqueClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *EstoqueClient) Baixar(ctx context.Context, requestID string, itens []ItemBaixaPayload) error {
	body, err := json.Marshal(baixarRequestPayload{Itens: itens})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/estoque/baixas", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		req.Header.Set(httpx.HeaderRequestID, requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ErrEstoqueIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	var erro httpx.ErroResposta
	_ = json.NewDecoder(resp.Body).Decode(&erro)
	if erro.Code == "SALDO_INSUFICIENTE" {
		return ErrSaldoInsuficiente
	}
	return ErrEstoqueIndisponivel
}
