package faturamento

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
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
