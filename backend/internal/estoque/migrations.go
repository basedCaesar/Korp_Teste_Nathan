package estoque

const migrationCreateProdutos = `
CREATE TABLE IF NOT EXISTS produtos (
    id BIGSERIAL PRIMARY KEY,
    codigo TEXT NOT NULL UNIQUE,
    descricao TEXT NOT NULL,
    saldo INTEGER NOT NULL DEFAULT 0 CHECK (saldo >= 0),
    version INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// Migrations lista, em ordem, os statements idempotentes do dominio estoque.
// Rodam no boot do servico (ver cmd/estoque/main.go), sem ferramenta externa.
var Migrations = []string{migrationCreateProdutos}
