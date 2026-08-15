package estoque

import "errors"

// Erros de dominio como sentinelas, classificados na camada HTTP via errors.Is.
var (
	ErrProdutoNaoEncontrado = errors.New("produto nao encontrado")
	ErrCodigoDuplicado      = errors.New("codigo ja cadastrado")

	// usados a partir do bloco 3 (baixa de estoque na impressao da nota)
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
	ErrConflitoVersao    = errors.New("conflito de versao no produto")
)
