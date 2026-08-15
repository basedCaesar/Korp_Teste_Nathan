package faturamento

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlInserirNota = `
		INSERT INTO notas DEFAULT VALUES
		RETURNING id, numero, status, created_at, updated_at`

	sqlBuscarNota = `
		SELECT id, numero, status, created_at, updated_at
		FROM notas WHERE id = $1`

	sqlListarNotas = `
		SELECT id, numero, status, created_at, updated_at
		FROM notas ORDER BY id`

	sqlExcluirNota = `
		DELETE FROM notas WHERE id = $1 AND status = 'ABERTA'`

	sqlListarItens = `
		SELECT id, nota_id, produto_id, produto_codigo, produto_descricao, quantidade, created_at
		FROM itens_nota WHERE nota_id = $1 ORDER BY id`

	sqlInserirItem = `
		INSERT INTO itens_nota (nota_id, produto_id, produto_codigo, produto_descricao, quantidade)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, nota_id, produto_id, produto_codigo, produto_descricao, quantidade, created_at`

	sqlAtualizarItem = `
		UPDATE itens_nota SET quantidade = $3
		WHERE id = $1 AND nota_id = $2
		RETURNING id, nota_id, produto_id, produto_codigo, produto_descricao, quantidade, created_at`

	sqlExcluirItem = `
		DELETE FROM itens_nota WHERE id = $1 AND nota_id = $2`

	sqlMarcarProcessando = `
		UPDATE notas SET status = 'PROCESSANDO', updated_at = now()
		WHERE id = $1 AND status = 'ABERTA'`

	sqlMarcarAberta = `
		UPDATE notas SET status = 'ABERTA', updated_at = now() WHERE id = $1`

	sqlMarcarFechada = `
		UPDATE notas SET status = 'FECHADA', updated_at = now() WHERE id = $1`

	sqlReabrirNotasTravadas = `
		UPDATE notas SET status = 'ABERTA', updated_at = now()
		WHERE status = 'PROCESSANDO' AND updated_at < now() - make_interval(secs => $1)`
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanNota(row pgx.Row) (Nota, error) {
	var n Nota
	err := row.Scan(&n.ID, &n.Numero, &n.Status, &n.CreatedAt, &n.UpdatedAt)
	return n, err
}

func scanItem(row pgx.Row) (Item, error) {
	var i Item
	err := row.Scan(&i.ID, &i.NotaID, &i.ProdutoID, &i.ProdutoCodigo, &i.ProdutoDescricao, &i.Quantidade, &i.CreatedAt)
	return i, err
}

func (r *Repository) CriarNota(ctx context.Context) (Nota, error) {
	row := r.pool.QueryRow(ctx, sqlInserirNota)
	return scanNota(row)
}

func (r *Repository) BuscarNota(ctx context.Context, id int64) (Nota, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarNota, id)
	n, err := scanNota(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Nota{}, ErrNotaNaoEncontrada
	}
	return n, err
}

func (r *Repository) ListarNotas(ctx context.Context) ([]Nota, error) {
	rows, err := r.pool.Query(ctx, sqlListarNotas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notas := []Nota{}
	for rows.Next() {
		n, err := scanNota(rows)
		if err != nil {
			return nil, err
		}
		notas = append(notas, n)
	}
	return notas, rows.Err()
}

func (r *Repository) ExcluirNota(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, sqlExcluirNota, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotaNaoEncontrada
	}
	return nil
}

func (r *Repository) ListarItens(ctx context.Context, notaID int64) ([]Item, error) {
	rows, err := r.pool.Query(ctx, sqlListarItens, notaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := []Item{}
	for rows.Next() {
		i, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		itens = append(itens, i)
	}
	return itens, rows.Err()
}

func (r *Repository) AdicionarItem(ctx context.Context, notaID, produtoID int64, produtoCodigo, produtoDescricao string, quantidade int) (Item, error) {
	row := r.pool.QueryRow(ctx, sqlInserirItem, notaID, produtoID, produtoCodigo, produtoDescricao, quantidade)
	return scanItem(row)
}

func (r *Repository) AtualizarItem(ctx context.Context, notaID, itemID int64, quantidade int) (Item, error) {
	row := r.pool.QueryRow(ctx, sqlAtualizarItem, itemID, notaID, quantidade)
	i, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrItemNaoEncontrado
	}
	return i, err
}

func (r *Repository) RemoverItem(ctx context.Context, notaID, itemID int64) error {
	tag, err := r.pool.Exec(ctx, sqlExcluirItem, itemID, notaID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrItemNaoEncontrado
	}
	return nil
}

func (r *Repository) MarcarProcessando(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, sqlMarcarProcessando, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	if _, err := r.BuscarNota(ctx, id); err != nil {
		return err
	}
	return ErrNotaNaoAberta
}

func (r *Repository) MarcarAberta(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, sqlMarcarAberta, id)
	return err
}

func (r *Repository) MarcarFechada(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, sqlMarcarFechada, id)
	return err
}

func (r *Repository) ReabrirNotasTravadas(ctx context.Context, limite time.Duration) (int64, error) {
	tag, err := r.pool.Exec(ctx, sqlReabrirNotasTravadas, limite.Seconds())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
