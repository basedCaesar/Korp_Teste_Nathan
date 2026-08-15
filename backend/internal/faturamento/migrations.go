package faturamento

const migrationCreateNotasSequence = `
CREATE SEQUENCE IF NOT EXISTS notas_numero_seq;
`

const migrationCreateNotas = `
CREATE TABLE IF NOT EXISTS notas (
    id BIGSERIAL PRIMARY KEY,
    numero BIGINT NOT NULL UNIQUE DEFAULT nextval('notas_numero_seq'),
    status TEXT NOT NULL DEFAULT 'ABERTA' CHECK (status IN ('ABERTA', 'PROCESSANDO', 'FECHADA')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

const migrationCreateItensNota = `
CREATE TABLE IF NOT EXISTS itens_nota (
    id BIGSERIAL PRIMARY KEY,
    nota_id BIGINT NOT NULL REFERENCES notas(id) ON DELETE CASCADE,
    produto_id BIGINT NOT NULL,
    produto_codigo TEXT NOT NULL,
    produto_descricao TEXT NOT NULL,
    quantidade INTEGER NOT NULL CHECK (quantidade > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

const migrationCreateOutbox = `
CREATE TABLE IF NOT EXISTS outbox (
    id BIGSERIAL PRIMARY KEY,
    nota_id BIGINT NOT NULL REFERENCES notas(id),
    destinatario TEXT NOT NULL,
    assunto TEXT NOT NULL,
    corpo TEXT NOT NULL,
    processado BOOLEAN NOT NULL DEFAULT false,
    tentativas INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);
`

var Migrations = []string{
	migrationCreateNotasSequence,
	migrationCreateNotas,
	migrationCreateItensNota,
	migrationCreateOutbox,
}
