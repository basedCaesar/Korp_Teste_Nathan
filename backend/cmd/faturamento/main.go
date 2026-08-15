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

	ctx := context.Background()
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, faturamento.Migrations); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	repo := faturamento.NewRepository(pool)
	svc := faturamento.NewService(repo)

	r := gin.New()
	r.Use(httpx.Recovery())
	r.Use(httpx.RequestIDMiddleware())
	httpx.RegisterHealth(r, "faturamento")
	faturamento.RegisterRoutes(r, svc)

	slog.Info("iniciando servico", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}
