package faturamento

import "time"

const (
	StatusAberta      = "ABERTA"
	StatusProcessando = "PROCESSANDO"
	StatusFechada     = "FECHADA"
)

type Nota struct {
	ID        int64     `json:"id"`
	Numero    int64     `json:"numero"`
	Status    string    `json:"status"`
	Itens     []Item    `json:"itens,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Item struct {
	ID               int64     `json:"id"`
	NotaID           int64     `json:"nota_id"`
	ProdutoID        int64     `json:"produto_id"`
	ProdutoCodigo    string    `json:"produto_codigo"`
	ProdutoDescricao string    `json:"produto_descricao"`
	Quantidade       int       `json:"quantidade"`
	CreatedAt        time.Time `json:"created_at"`
}
