package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"korp/internal/db"
	"korp/internal/faturamento"
	"korp/internal/httpx"
	"korp/internal/mailer"
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
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	if smtpHost == "" || smtpPort == "" {
		slog.Error("SMTP_HOST/SMTP_PORT nao configurados")
		os.Exit(1)
	}
	mailFrom := os.Getenv("MAIL_FROM")
	if mailFrom == "" {
		mailFrom = "noreply@korp.local"
	}
	notaConfirmacaoEmail := os.Getenv("NOTA_CONFIRMACAO_EMAIL")
	if notaConfirmacaoEmail == "" {
		notaConfirmacaoEmail = "cliente@korp.local"
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
	svc := faturamento.NewService(repo, estoqueClient, notaConfirmacaoEmail)
	idemStore := httpx.NewIdempotencyStore(pool)
	mail := mailer.New(smtpHost, smtpPort, mailFrom)

	go faturamento.RunReaper(ctx, repo)
	go faturamento.RunOutboxConsumer(ctx, repo, mail)

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
