package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"korp/internal/db"
	"korp/internal/faturamento"
	"korp/internal/httpx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "faturamento")
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL nao configurada")
		os.Exit(1)
	}
	estoqueURL := os.Getenv("ESTOQUE_URL")
	if estoqueURL == "" {
		slog.Error("ESTOQUE_URL nao configurada")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrations := append([]string{httpx.MigrationCreateIdempotencyKeys}, faturamento.Migrations...)
	if err := db.RunMigrations(ctx, pool, migrations); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	repo := faturamento.NewRepository(pool)
	estoqueClient := faturamento.NewEstoqueClient(estoqueURL)
	svc := faturamento.NewService(repo, estoqueClient)
	idemStore := httpx.NewIdempotencyStore(pool)

	go faturamento.RunReaper(ctx, repo)

	r := gin.New()
	r.Use(httpx.Recovery())
	r.Use(httpx.RequestIDMiddleware())
	httpx.RegisterHealth(r, "faturamento")
	faturamento.RegisterRoutes(r, svc, idemStore)

	slog.Info("iniciando servico", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}
