package faturamento

import (
	"context"
	"log/slog"
	"time"

	"korp/internal/mailer"
)

const (
	OutboxIntervalo     = 5 * time.Second
	outboxMaxTentativas = 5
)

func RunOutboxConsumer(ctx context.Context, repo *Repository, mail *mailer.Mailer) {
	ticker := time.NewTicker(OutboxIntervalo)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processarOutboxPendente(ctx, repo, mail)
		}
	}
}

func processarOutboxPendente(ctx context.Context, repo *Repository, mail *mailer.Mailer) {
	itens, err := repo.ListarOutboxPendente(ctx, outboxMaxTentativas)
	if err != nil {
		slog.Error("outbox: falha ao listar pendentes", "error", err)
		return
	}

	for _, item := range itens {
		if err := mail.Enviar(item.Destinatario, item.Assunto, item.Corpo); err != nil {
			slog.Error("outbox: falha ao enviar email", "error", err, "outbox_id", item.ID, "tentativas", item.Tentativas)
			if err := repo.IncrementarTentativaOutbox(ctx, item.ID); err != nil {
				slog.Error("outbox: falha ao incrementar tentativas", "error", err, "outbox_id", item.ID)
			}
			continue
		}
		if err := repo.MarcarOutboxProcessado(ctx, item.ID); err != nil {
			slog.Error("outbox: falha ao marcar processado", "error", err, "outbox_id", item.ID)
			continue
		}
		slog.Info("outbox: email enviado", "outbox_id", item.ID, "nota_id", item.NotaID)
	}
}
