package faturamento

import "errors"

var (
	ErrNotaNaoEncontrada  = errors.New("nota nao encontrada")
	ErrItemNaoEncontrado  = errors.New("item nao encontrado")
	ErrNotaNaoAberta       = errors.New("nota nao esta aberta")
	ErrEstoqueIndisponivel = errors.New("estoque indisponivel")
	ErrSaldoInsuficiente   = errors.New("saldo insuficiente")
)
