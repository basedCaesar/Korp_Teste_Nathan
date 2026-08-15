package estoque

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeRow struct {
	produto Produto
}

func (r fakeRow) Scan(dest ...any) error {
	*dest[0].(*int64) = r.produto.ID
	*dest[1].(*string) = r.produto.Codigo
	*dest[2].(*string) = r.produto.Descricao
	*dest[3].(*int) = r.produto.Saldo
	*dest[4].(*string) = r.produto.Categoria
	*dest[5].(*int) = r.produto.Version
	*dest[6].(*time.Time) = r.produto.CreatedAt
	*dest[7].(*time.Time) = r.produto.UpdatedAt
	return nil
}

type fakeDB struct {
	produtos   []Produto
	produtoIdx int
	execTags   []pgconn.CommandTag
	execIdx    int
}

func (f *fakeDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	p := f.produtos[f.produtoIdx]
	if f.produtoIdx < len(f.produtos)-1 {
		f.produtoIdx++
	}
	return fakeRow{produto: p}
}

func (f *fakeDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag := f.execTags[f.execIdx]
	if f.execIdx < len(f.execTags)-1 {
		f.execIdx++
	}
	return tag, nil
}

func produtoTeste(saldo, version int) Produto {
	agora := time.Now()
	return Produto{ID: 1, Codigo: "P001", Descricao: "Teste", Saldo: saldo, Version: version, CreatedAt: agora, UpdatedAt: agora}
}

func TestBaixarProduto_Sucesso(t *testing.T) {
	db := &fakeDB{
		produtos: []Produto{produtoTeste(10, 0)},
		execTags: []pgconn.CommandTag{pgconn.NewCommandTag("UPDATE 1")},
	}

	err := baixarProduto(context.Background(), db, 1, 2)

	if err != nil {
		t.Fatalf("esperava sucesso, veio erro: %v", err)
	}
	if db.execIdx != 0 {
		t.Fatalf("esperava 1 tentativa de UPDATE, teve %d", db.execIdx+1)
	}
}

func TestBaixarProduto_ConflitoDeVersaoRetentaComSucesso(t *testing.T) {
	db := &fakeDB{
		produtos: []Produto{produtoTeste(10, 0), produtoTeste(10, 1)},
		execTags: []pgconn.CommandTag{pgconn.NewCommandTag("UPDATE 0"), pgconn.NewCommandTag("UPDATE 1")},
	}

	err := baixarProduto(context.Background(), db, 1, 2)

	if err != nil {
		t.Fatalf("esperava sucesso apos retry, veio erro: %v", err)
	}
	if db.execIdx != 1 {
		t.Fatalf("esperava 2 tentativas de UPDATE, teve %d", db.execIdx+1)
	}
}

func TestBaixarProduto_SaldoInsuficiente(t *testing.T) {
	db := &fakeDB{
		produtos: []Produto{produtoTeste(2, 0), produtoTeste(2, 0)},
		execTags: []pgconn.CommandTag{pgconn.NewCommandTag("UPDATE 0")},
	}

	err := baixarProduto(context.Background(), db, 1, 5)

	if !errors.Is(err, ErrSaldoInsuficiente) {
		t.Fatalf("esperava ErrSaldoInsuficiente, veio: %v", err)
	}
}

func TestBaixarProduto_ConflitoVersaoEsgotaTentativas(t *testing.T) {
	db := &fakeDB{
		produtos: []Produto{
			produtoTeste(100, 0),
			produtoTeste(100, 1),
			produtoTeste(100, 2),
			produtoTeste(100, 3),
		},
		execTags: []pgconn.CommandTag{
			pgconn.NewCommandTag("UPDATE 0"),
			pgconn.NewCommandTag("UPDATE 0"),
			pgconn.NewCommandTag("UPDATE 0"),
		},
	}

	err := baixarProduto(context.Background(), db, 1, 2)

	if !errors.Is(err, ErrConflitoVersao) {
		t.Fatalf("esperava ErrConflitoVersao apos esgotar tentativas, veio: %v", err)
	}
	if db.execIdx != maxTentativasBaixa-1 {
		t.Fatalf("esperava %d tentativas de UPDATE, teve %d", maxTentativasBaixa, db.execIdx+1)
	}
}
