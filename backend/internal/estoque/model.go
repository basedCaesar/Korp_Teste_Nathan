package estoque

import "time"

// Produto e um item de estoque. Version suporta o lock otimista usado na baixa (bloco 3).
type ItemBaixa struct {
	ProdutoID  int64 `json:"produto_id"`
	Quantidade int   `json:"quantidade"`
}

type Produto struct {
	ID        int64     `json:"id"`
	Codigo    string    `json:"codigo"`
	Descricao string    `json:"descricao"`
	Saldo     int       `json:"saldo"`
	Categoria string    `json:"categoria"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
