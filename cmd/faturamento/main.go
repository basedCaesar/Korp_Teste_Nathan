package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"korp/internal/httpx"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "faturamento")
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	r := gin.New()
	r.Use(gin.Recovery())
	httpx.RegisterHealth(r, "faturamento")

	slog.Info("iniciando servico", "port", port)
	if err := r.Run(":" + port); err != nil {
		slog.Error("falha ao iniciar servidor", "error", err)
		os.Exit(1)
	}
}
