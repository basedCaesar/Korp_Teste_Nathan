package estoque

import "context"

// Service e a camada fina entre o handler HTTP e o repository.
// A partir do bloco 3 tambem concentra a logica de baixa com lock otimista.
type Service struct {
	repo     *Repository
	iaClient *IAClient
}

func NewService(repo *Repository, iaClient *IAClient) *Service {
	return &Service{repo: repo, iaClient: iaClient}
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

func (s *Service) BaixarItens(ctx context.Context, itens []ItemBaixa) error {
	return s.repo.BaixarItens(ctx, itens)
}

func (s *Service) Sugerir(ctx context.Context, codigo string) (SugestaoResponse, error) {
	catalogo, err := s.repo.Listar(ctx)
	if err != nil {
		return SugestaoResponse{}, err
	}

	descricao, similaresCodigos, err := s.iaClient.Sugerir(ctx, codigo, catalogo)
	if err != nil {
		return SugestaoResponse{}, err
	}

	produtosSimilares := make([]Produto, 0, len(similaresCodigos))
	for _, cod := range similaresCodigos {
		for _, p := range catalogo {
			if p.Codigo == cod {
				produtosSimilares = append(produtosSimilares, p)
				break
			}
		}
	}

	return SugestaoResponse{
		Codigo:            codigo,
		DescricaoSugerida: descricao,
		ProdutosSimilares: produtosSimilares,
	}, nil
}
