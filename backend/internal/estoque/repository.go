package estoque

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL em const, nunca montado por concatenacao.
const (
	sqlInserirProduto = `
		INSERT INTO produtos (codigo, descricao, saldo)
		VALUES ($1, $2, $3)
		RETURNING id, codigo, descricao, saldo, version, created_at, updated_at`

	sqlBuscarProduto = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos WHERE id = $1`

	sqlListarProdutos = `
		SELECT id, codigo, descricao, saldo, version, created_at, updated_at
		FROM produtos ORDER BY id`

	sqlAtualizarProduto = `
		UPDATE produtos
		SET descricao = $2, saldo = $3, version = version + 1, updated_at = now()
		WHERE id = $1
		RETURNING id, codigo, descricao, saldo, version, created_at, updated_at`

	sqlExcluirProduto = `DELETE FROM produtos WHERE id = $1`
)

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

func (r *Repository) Criar(ctx context.Context, codigo, descricao string, saldo int) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlInserirProduto, codigo, descricao, saldo)
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

func (r *Repository) BuscarPorID(ctx context.Context, id int64) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarProduto, id)
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

func (r *Repository) Atualizar(ctx context.Context, id int64, descricao string, saldo int) (Produto, error) {
	row := r.pool.QueryRow(ctx, sqlAtualizarProduto, id, descricao, saldo)
	p, err := scanProduto(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Produto{}, ErrProdutoNaoEncontrado
	}
	return p, err
}

func (r *Repository) Excluir(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, sqlExcluirProduto, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProdutoNaoEncontrado
	}
	return nil
}
