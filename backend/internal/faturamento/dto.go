package faturamento

type AdicionarItemRequest struct {
	ProdutoID        int64  `json:"produto_id" binding:"required"`
	ProdutoCodigo    string `json:"produto_codigo" binding:"required"`
	ProdutoDescricao string `json:"produto_descricao" binding:"required"`
	Quantidade       int    `json:"quantidade" binding:"required,gt=0"`
}

type AtualizarItemRequest struct {
	Quantidade int `json:"quantidade" binding:"required,gt=0"`
}
