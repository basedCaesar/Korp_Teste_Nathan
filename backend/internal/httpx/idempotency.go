package httpx

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const HeaderIdempotencyKey = "Idempotency-Key"

const MigrationCreateIdempotencyKeys = `
CREATE TABLE IF NOT EXISTS idempotency_keys (
    chave TEXT PRIMARY KEY,
    status_code INTEGER NOT NULL,
    corpo BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

const (
	sqlBuscarIdempotencia = `SELECT status_code, corpo FROM idempotency_keys WHERE chave = $1`
	sqlSalvarIdempotencia = `INSERT INTO idempotency_keys (chave, status_code, corpo) VALUES ($1, $2, $3) ON CONFLICT (chave) DO NOTHING`
)

type IdempotencyStore struct {
	pool *pgxpool.Pool
}

func NewIdempotencyStore(pool *pgxpool.Pool) *IdempotencyStore {
	return &IdempotencyStore{pool: pool}
}

func (s *IdempotencyStore) buscar(ctx context.Context, chave string) (int, []byte, bool, error) {
	var status int
	var corpo []byte
	err := s.pool.QueryRow(ctx, sqlBuscarIdempotencia, chave).Scan(&status, &corpo)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil, false, nil
	}
	if err != nil {
		return 0, nil, false, err
	}
	return status, corpo, true, nil
}

func (s *IdempotencyStore) salvar(ctx context.Context, chave string, status int, corpo []byte) error {
	if corpo == nil {
		corpo = []byte{}
	}
	_, err := s.pool.Exec(ctx, sqlSalvarIdempotencia, chave, status, corpo)
	return err
}

type responseCapture struct {
	gin.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (w *responseCapture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseCapture) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Idempotency(store *IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		chave := c.GetHeader(HeaderIdempotencyKey)
		if chave == "" {
			RespondError(c, http.StatusBadRequest, "IDEMPOTENCY_KEY_OBRIGATORIA", "header Idempotency-Key e obrigatorio")
			return
		}

		status, corpo, achou, err := store.buscar(c.Request.Context(), chave)
		if err == nil && achou {
			c.Data(status, "application/json", corpo)
			c.Abort()
			return
		}

		capture := &responseCapture{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = capture
		c.Next()

		if capture.status > 0 && capture.status < http.StatusInternalServerError {
			if err := store.salvar(c.Request.Context(), chave, capture.status, capture.buf.Bytes()); err != nil {
				slog.Error("falha ao salvar chave de idempotencia", "error", err, "chave", chave)
			}
		}
	}
}
