// Package db cuida do pool de conexoes pgx e das migrations idempotentes
// rodadas no boot de cada servico (sem ferramenta externa de migration).
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool cria e testa o pool de conexoes a partir da DATABASE_URL.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("criar pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping banco: %w", err)
	}
	return pool, nil
}

// RunMigrations executa, em ordem, statements idempotentes (CREATE TABLE IF NOT EXISTS ...).
// Cada servico roda apenas as migrations do seu proprio dominio no boot.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, statements []string) error {
	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("executar migration: %w", err)
		}
	}
	return nil
}
