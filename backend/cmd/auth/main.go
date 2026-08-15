package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"korp/internal/auth"
	"korp/internal/db"
	"korp/internal/httpx"
	"korp/internal/mailer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "auth")
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL nao configurada")
		os.Exit(1)
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET nao configurada")
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
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		slog.Error("falha ao conectar no banco", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, auth.Migrations); err != nil {
		slog.Error("falha ao rodar migrations", "error", err)
		os.Exit(1)
	}

	repo := auth.NewRepository(pool)
	mail := mailer.New(smtpHost, smtpPort, mailFrom)
	svc := auth.NewService(repo, mail, jwtSecret, baseURL)

	r := gin.New()
	r.Use(httpx.Recovery())
	r.Use(httpx.RequestIDMiddleware())
	httpx.RegisterHealth(r, "auth")
	auth.RegisterRoutes(r, svc, jwtSecret)

	slog.Info("iniciando servico", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}
