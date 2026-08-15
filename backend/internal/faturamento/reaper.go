package faturamento

import (
	"context"
	"log/slog"
	"time"
)

const (
	ReaperIntervalo = 30 * time.Second
	ReaperLimite    = 2 * time.Minute
)

func RunReaper(ctx context.Context, repo *Repository) {
	ticker := time.NewTicker(ReaperIntervalo)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := repo.ReabrirNotasTravadas(ctx, ReaperLimite)
			if err != nil {
				slog.Error("reaper falhou ao reabrir notas travadas", "error", err)
				continue
			}
			if n > 0 {
				slog.Info("reaper reabriu notas travadas", "quantidade", n)
			}
		}
	}
}
