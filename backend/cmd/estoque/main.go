package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"korp/internal/db"
	"korp/internal/estoque"
	"korp/internal/httpx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "estoque")
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL nao configurada")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrations := append([]string{httpx.MigrationCreateIdempotencyKeys}, estoque.Migrations...)
	if err := db.RunMigrations(ctx, pool, migrations); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	repo := estoque.NewRepository(pool)
	svc := estoque.NewService(repo)
	idemStore := httpx.NewIdempotencyStore(pool)

	r := gin.New()
	r.Use(httpx.Recovery())
	r.Use(httpx.RequestIDMiddleware())
	httpx.RegisterHealth(r, "estoque")
	estoque.RegisterRoutes(r, svc, idemStore)

	slog.Info("iniciando servico", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}
