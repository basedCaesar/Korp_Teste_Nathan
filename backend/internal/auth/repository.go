package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	sqlInserirUsuario = `
		INSERT INTO users (email, senha_hash, token_verificacao, token_verificacao_expira_em)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, senha_hash, verificado, token_verificacao, token_verificacao_expira_em, created_at, updated_at`

	sqlBuscarPorEmail = `
		SELECT id, email, senha_hash, verificado, token_verificacao, token_verificacao_expira_em, created_at, updated_at
		FROM users WHERE email = $1`

	sqlBuscarPorToken = `
		SELECT id, email, senha_hash, verificado, token_verificacao, token_verificacao_expira_em, created_at, updated_at
		FROM users WHERE token_verificacao = $1`

	sqlBuscarPorID = `
		SELECT id, email, senha_hash, verificado, token_verificacao, token_verificacao_expira_em, created_at, updated_at
		FROM users WHERE id = $1`

	sqlMarcarVerificado = `
		UPDATE users SET verificado = true, token_verificacao = NULL, token_verificacao_expira_em = NULL, updated_at = now()
		WHERE id = $1`
)

const pgErrCodeUniqueViolation = "23505"

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.SenhaHash, &u.Verificado, &u.TokenVerificacao, &u.TokenVerificacaoExpiraEm, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (r *Repository) Criar(ctx context.Context, email, senhaHash, token string, expiraEm time.Time) (User, error) {
	row := r.pool.QueryRow(ctx, sqlInserirUsuario, email, senhaHash, token, expiraEm)
	u, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrCodeUniqueViolation {
			return User{}, ErrEmailJaCadastrado
		}
		return User{}, err
	}
	return u, nil
}

func (r *Repository) BuscarPorEmail(ctx context.Context, email string) (User, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarPorEmail, email)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrCredenciaisInvalidas
	}
	return u, err
}

func (r *Repository) BuscarPorToken(ctx context.Context, token string) (User, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarPorToken, token)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrTokenInvalido
	}
	return u, err
}

func (r *Repository) BuscarPorID(ctx context.Context, id int64) (User, error) {
	row := r.pool.QueryRow(ctx, sqlBuscarPorID, id)
	return scanUser(row)
}

func (r *Repository) MarcarVerificado(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, sqlMarcarVerificado, id)
	return err
}
