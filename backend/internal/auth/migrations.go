package auth

const migrationCreateUsers = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    senha_hash TEXT NOT NULL,
    verificado BOOLEAN NOT NULL DEFAULT false,
    token_verificacao TEXT,
    token_verificacao_expira_em TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

var Migrations = []string{migrationCreateUsers}
