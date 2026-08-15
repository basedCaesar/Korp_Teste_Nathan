package faturamento

import (
	"context"
	"fmt"
)

type Service struct {
	repo             *Repository
	estoqueClient    *EstoqueClient
	notificacaoEmail string
}

func NewService(repo *Repository, estoqueClient *EstoqueClient, notificacaoEmail string) *Service {
	return &Service{repo: repo, estoqueClient: estoqueClient, notificacaoEmail: notificacaoEmail}
}

func (s *Service) Criar(ctx context.Context, userID int64) (Nota, error) {
	return s.repo.CriarNota(ctx, userID)
}

func (s *Service) Listar(ctx context.Context, userID int64) ([]Nota, error) {
	return s.repo.ListarNotas(ctx, userID)
}

func (s *Service) Buscar(ctx context.Context, id, userID int64) (Nota, error) {
	nota, err := s.repo.BuscarNota(ctx, id, userID)
	if err != nil {
		return Nota{}, err
	}
	itens, err := s.repo.ListarItens(ctx, id)
	if err != nil {
		return Nota{}, err
	}
	nota.Itens = itens
	return nota, nil
}

func (s *Service) Excluir(ctx context.Context, id, userID int64) error {
	nota, err := s.repo.BuscarNota(ctx, id, userID)
	if err != nil {
		return err
	}
	if nota.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	return s.repo.ExcluirNota(ctx, id, userID)
}

func (s *Service) AdicionarItem(ctx context.Context, authHeader string, notaID, userID, produtoID int64, produtoCodigo, produtoDescricao string, quantidade int) (Item, error) {
	nota, err := s.repo.BuscarNota(ctx, notaID, userID)
	if err != nil {
		return Item{}, err
	}
	if nota.Status != StatusAberta {
		return Item{}, ErrNotaNaoAberta
	}
	if err := s.estoqueClient.BuscarProduto(ctx, produtoID, authHeader); err != nil {
		return Item{}, err
	}
	return s.repo.AdicionarItem(ctx, notaID, produtoID, produtoCodigo, produtoDescricao, quantidade)
}

func (s *Service) AtualizarItem(ctx context.Context, notaID, itemID, userID int64, quantidade int) (Item, error) {
	nota, err := s.repo.BuscarNota(ctx, notaID, userID)
	if err != nil {
		return Item{}, err
	}
	if nota.Status != StatusAberta {
		return Item{}, ErrNotaNaoAberta
	}
	return s.repo.AtualizarItem(ctx, notaID, itemID, quantidade)
}

func (s *Service) RemoverItem(ctx context.Context, notaID, itemID, userID int64) error {
	nota, err := s.repo.BuscarNota(ctx, notaID, userID)
	if err != nil {
		return err
	}
	if nota.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	return s.repo.RemoverItem(ctx, notaID, itemID)
}

func (s *Service) Imprimir(ctx context.Context, requestID string, notaID, userID int64) error {
	if err := s.repo.MarcarProcessando(ctx, notaID, userID); err != nil {
		return err
	}

	itens, err := s.repo.ListarItens(ctx, notaID)
	if err != nil {
		_ = s.repo.MarcarAberta(ctx, notaID)
		return err
	}

	payload := make([]ItemBaixaPayload, len(itens))
	for i, item := range itens {
		payload[i] = ItemBaixaPayload{ProdutoID: item.ProdutoID, Quantidade: item.Quantidade}
	}

	if err := s.estoqueClient.Baixar(ctx, requestID, payload); err != nil {
		_ = s.repo.MarcarAberta(ctx, notaID)
		return err
	}

	assunto := "Nota fiscal emitida"
	corpo := fmt.Sprintf("A nota fiscal #%d foi emitida com sucesso.", notaID)
	return s.repo.FecharNotaComOutbox(ctx, notaID, s.notificacaoEmail, assunto, corpo)
}
