package estoque

import "context"

// Service e a camada fina entre o handler HTTP e o repository.
// A partir do bloco 3 tambem concentra a logica de baixa com lock otimista.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Criar(ctx context.Context, codigo, descricao string, saldo int) (Produto, error) {
	return s.repo.Criar(ctx, codigo, descricao, saldo)
}

func (s *Service) Buscar(ctx context.Context, id int64) (Produto, error) {
	return s.repo.BuscarPorID(ctx, id)
}

func (s *Service) Listar(ctx context.Context) ([]Produto, error) {
	return s.repo.Listar(ctx)
}

func (s *Service) Atualizar(ctx context.Context, id int64, descricao string, saldo int) (Produto, error) {
	return s.repo.Atualizar(ctx, id, descricao, saldo)
}

func (s *Service) Excluir(ctx context.Context, id int64) error {
	return s.repo.Excluir(ctx, id)
}
