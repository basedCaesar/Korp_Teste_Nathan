package estoque

import (
	"context"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL em const, nunca montado por concatenacao.
const (
	sqlInserirProduto = `
		INSERT INTO produtos (codigo, descricao, saldo, user_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, codigo, descricao, saldo, version, created_at, updated_at`

	sqlBuscarProduto = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos WHERE id = $1`

	sqlBuscarProdutoDoUsuario = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos WHERE id = $1 AND user_id = $2`

	sqlListarProdutos = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos ORDER BY id`

	sqlListarProdutosDoUsuario = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos WHERE user_id = $1 ORDER BY id`

	sqlAtualizarProduto = `
		UPDATE produtos
		SET descricao = $2, saldo = $3, version = version + 1, updated_at = now()
		WHERE id = $1 AND user_id = $4
		RETURNING id, codigo, descricao, saldo, version, created_at, updated_at`

	sqlExcluirProduto = `DELETE FROM produtos WHERE id = $1 AND user_id = $2`

	sqlBaixarProduto = `
		UPDATE produtos SET saldo = saldo - $2, version = version + 1, updated_at = now()
		WHERE id = $1 AND version = $3 AND saldo >= $2`
)

const maxTentativasBaixa = 3

type dbtx interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

const pgErrCodeUniqueViolation = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanProduto(row pgx.Row) (Produto, error) {
	var p Produto
	err := row.Scan(&p.ID, &p.Codigo, &p.Descricao, &p.Saldo, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) Criar(ctx context.Context, userID int64, codigo, descricao string, saldo int) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlInserirProduto, codigo, descricao, saldo, userID)
	p, err := scanProduto(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
			return Produto{}, ErrCodigoDuplicado
		}
		return Produto{}, err
	}
	return p, nil
}

func buscarProduto(ctx context.Context, db dbtx, id int64) (Produto, error) {
	row := db.QueryRow(ctx, sqlBuscarProduto, id)
	p, err := scanProduto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Produto{}, ErrProdutoNaoEncontrado
	}
	return p, err
}

func (r *Repository) BuscarPorID(ctx context.Context, id, userID int64) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarProdutoDoUsuario, id, userID)
	p, err := scanProduto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Produto{}, ErrProdutoNaoEncontrado
	}
	return p, err
}

func (r *Repository) Listar(ctx context.Context) ([]Produto, error) {
	rows, err := r.pool.Query(ctx, sqlListarProdutos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	produtos := []Produto{}
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, err
		}
		produtos = append(produtos, p)
	}
	return produtos, rows.Err()
}

func (r *Repository) ListarPorUsuario(ctx context.Context, userID int64) ([]Produto, error) {
	rows, err := r.pool.Query(ctx, sqlListarProdutosDoUsuario, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	produtos := []Produto{}
	for rows.Next() {
		p, err := scanProduto(rows)
		if err != nil {
			return nil, err
		}
		produtos = append(produtos, p)
	}
	return produtos, rows.Err()
}

func (r *Repository) Atualizar(ctx context.Context, id, userID int64, descricao string, saldo int) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlAtualizarProduto, id, descricao, saldo, userID)
	p, err := scanProduto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Produto{}, ErrProdutoNaoEncontrado
	}
	return p, err
}

func (r *Repository) Excluir(ctx context.Context, id, userID int64) error {
	tag, err := r.pool.Exec(ctx, sqlExcluirProduto, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProdutoNaoEncontrado
	}
	return nil
}

func baixarProduto(ctx context.Context, db dbtx, produtoID int64, quantidade int) error {
	produto, err := buscarProduto(ctx, db, produtoID)
	if err != nil {
		return err
	}
	for tentativa := 1; tentativa <= maxTentativasBaixa; tentativa++ {
		tag, err := db.Exec(ctx, sqlBaixarProduto, produtoID, quantidade, produto.Version)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			return nil
		}
		produto, err = buscarProduto(ctx, db, produtoID)
		if err != nil {
			return err
		}
		if produto.Saldo < quantidade {
			return ErrSaldoInsuficiente
		}
	}
	return ErrConflitoVersao
}

func (r *Repository) BaixarItens(ctx context.Context, itens []ItemBaixa) error {
	ordenados := make([]ItemBaixa, len(itens))
	copy(ordenados, itens)
	sort.Slice(ordenados, func(i, j int) bool { return ordenados[i].ProdutoID < ordenados[j].ProdutoID })

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range ordenados {
		if err := baixarProduto(ctx, tx, item.ProdutoID, item.Quantidade); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
