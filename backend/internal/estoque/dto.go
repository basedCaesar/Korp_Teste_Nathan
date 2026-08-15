package estoque

// CriarProdutoRequest e o corpo esperado em POST /produtos.
type CriarProdutoRequest struct {
	Codigo    string `json:"codigo" binding:"required"`
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
}

// AtualizarProdutoRequest e o corpo esperado em PUT /produtos/:id.
type AtualizarProdutoRequest struct {
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
}
