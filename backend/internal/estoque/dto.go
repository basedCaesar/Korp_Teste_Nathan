package estoque

// CriarProdutoRequest e o corpo esperado em POST /produtos.
type CriarProdutoRequest struct {
	Codigo    string `json:"codigo" binding:"required"`
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
	Categoria string `json:"categoria"`
}

// AtualizarProdutoRequest e o corpo esperado em PUT /produtos/:id.
type AtualizarProdutoRequest struct {
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
	Categoria string `json:"categoria"`
}

type BaixarRequest struct {
	Itens []ItemBaixaRequest `json:"itens" binding:"required,min=1,dive"`
}

type ItemBaixaRequest struct {
	ProdutoID  int64 `json:"produto_id" binding:"required"`
	Quantidade int   `json:"quantidade" binding:"required,gt=0"`
}

type SugestaoRequest struct {
	Codigo    string `json:"codigo" binding:"required"`
	Categoria string `json:"categoria"`
}

type SugestaoResponse struct {
	Codigo             string    `json:"codigo"`
	DescricaoSugerida  string    `json:"descricao_sugerida"`
	ProdutosSimilares  []Produto `json:"produtos_similares"`
}
