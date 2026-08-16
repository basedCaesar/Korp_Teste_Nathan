package faturamento

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sony/gobreaker"

	"korp/internal/httpx"
	"korp/internal/resilience"
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
	breaker    *gobreaker.CircuitBreaker
}

func NewEstoqueClient(baseURL string) *EstoqueClient {
	return &EstoqueClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		breaker:    resilience.NewBreaker("estoque"),
	}
}

var errTransporte = errors.New("falha de transporte ao chamar estoque")

func (c *EstoqueClient) Baixar(ctx context.Context, requestID string, itens []ItemBaixaPayload) error {
	var resultado error

	tentar := func(ctx context.Context) error {
		status, corpo, err := c.enviarBaixa(ctx, requestID, itens)
		if err != nil {
			return errTransporte
		}
		if status >= http.StatusInternalServerError {
			return errTransporte
		}
		if status == http.StatusOK {
			resultado = nil
			return nil
		}
		resultado = classificarErroNegocio(corpo)
		return nil
	}

	_, err := c.breaker.Execute(func() (any, error) {
		return nil, resilience.Retry(ctx, 3, 200*time.Millisecond, func(err error) bool {
			return errors.Is(err, errTransporte)
		}, tentar)
	})
	if err != nil {
		return ErrEstoqueIndisponivel
	}
	return resultado
}

func (c *EstoqueClient) enviarBaixa(ctx context.Context, requestID string, itens []ItemBaixaPayload) (int, []byte, error) {
	body, err := json.Marshal(baixarRequestPayload{Itens: itens})
	if err != nil {
		return 0, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/estoque/baixas", bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if requestID != "" {
		req.Header.Set(httpx.HeaderRequestID, requestID)
		req.Header.Set(httpx.HeaderIdempotencyKey, requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, corpo, nil
}

// BuscarProduto valida que o produto existe e pertence ao dono do token repassado, antes de um
// item ser gravado numa nota. Sem retry/circuit breaker aqui: e' uma checagem de validacao, nao
// a baixa em si, chamada sincrona simples com timeout curto.
func (c *EstoqueClient) BuscarProduto(ctx context.Context, produtoID int64, authHeader string) error {
	url := fmt.Sprintf("%s/produtos/%d", c.baseURL, produtoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ErrEstoqueIndisponivel
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized {
		return ErrProdutoInvalido
	}
	return ErrEstoqueIndisponivel
}

func (c *EstoqueClient) Ping(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func classificarErroNegocio(corpo []byte) error {
	var erro httpx.ErroResposta
	_ = json.Unmarshal(corpo, &erro)
	if erro.Code == "SALDO_INSUFICIENTE" {
		return ErrSaldoInsuficiente
	}
	return ErrEstoqueIndisponivel
}
