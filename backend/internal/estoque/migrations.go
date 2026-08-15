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

const migrationAddProdutosUserID = `
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS user_id BIGINT NOT NULL DEFAULT 0;
`

const migrationIndexProdutosUserID = `
CREATE INDEX IF NOT EXISTS idx_produtos_user_id ON produtos(user_id);
`

const migrationAddProdutosCategoria = `
ALTER TABLE produtos ADD COLUMN IF NOT EXISTS categoria TEXT;
`

// Migrations lista, em ordem, os statements idempotentes do dominio estoque.
// Rodam no boot do servico (ver cmd/estoque/main.go), sem ferramenta externa.
var Migrations = []string{
	migrationCreateProdutos,
	migrationAddProdutosUserID,
	migrationIndexProdutosUserID,
	migrationAddProdutosCategoria,
}
