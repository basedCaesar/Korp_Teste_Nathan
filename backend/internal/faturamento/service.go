package faturamento

import "context"

type Service struct {
	repo          *Repository
	estoqueClient *EstoqueClient
}

func NewService(repo *Repository, estoqueClient *EstoqueClient) *Service {
	return &Service{repo: repo, estoqueClient: estoqueClient}
}

func (s *Service) Criar(ctx context.Context) (Nota, error) {
	return s.repo.CriarNota(ctx)
}

func (s *Service) Listar(ctx context.Context) ([]Nota, error) {
	return s.repo.ListarNotas(ctx)
}

func (s *Service) Buscar(ctx context.Context, id int64) (Nota, error) {
	nota, err := s.repo.BuscarNota(ctx, id)
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

func (s *Service) Excluir(ctx context.Context, id int64) error {
	nota, err := s.repo.BuscarNota(ctx, id)
	if err != nil {
		return err
	}
	if nota.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	return s.repo.ExcluirNota(ctx, id)
}

func (s *Service) AdicionarItem(ctx context.Context, notaID, produtoID int64, produtoCodigo, produtoDescricao string, quantidade int) (Item, error) {
	nota, err := s.repo.BuscarNota(ctx, notaID)
	if err != nil {
		return Item{}, err
	}
	if nota.Status != StatusAberta {
		return Item{}, ErrNotaNaoAberta
	}
	return s.repo.AdicionarItem(ctx, notaID, produtoID, produtoCodigo, produtoDescricao, quantidade)
}

func (s *Service) AtualizarItem(ctx context.Context, notaID, itemID int64, quantidade int) (Item, error) {
	nota, err := s.repo.BuscarNota(ctx, notaID)
	if err != nil {
		return Item{}, err
	}
	if nota.Status != StatusAberta {
		return Item{}, ErrNotaNaoAberta
	}
	return s.repo.AtualizarItem(ctx, notaID, itemID, quantidade)
}

func (s *Service) RemoverItem(ctx context.Context, notaID, itemID int64) error {
	nota, err := s.repo.BuscarNota(ctx, notaID)
	if err != nil {
		return err
	}
	if nota.Status != StatusAberta {
		return ErrNotaNaoAberta
	}
	return s.repo.RemoverItem(ctx, notaID, itemID)
}

func (s *Service) Imprimir(ctx context.Context, requestID string, notaID int64) error {
	if err := s.repo.MarcarProcessando(ctx, notaID); err != nil {
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

	return s.repo.MarcarFechada(ctx, notaID)
}
