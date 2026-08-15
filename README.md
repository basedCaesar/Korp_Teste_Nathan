# Korp Teste — Sistema de Emissão de Notas Fiscais

Teste técnico: sistema de emissão de notas fiscais em arquitetura de microsserviços, backend em
Go e frontend em Angular.

## Arquitetura

Três serviços independentes, cada um com seu próprio banco Postgres, mais um SMTP fake (MailHog)
pra capturar e-mails sem depender de conta real:

| Serviço | Porta | Database | Responsabilidade |
|---|---|---|---|
| `estoque` | 8082 | `estoque_db` | produtos, saldos, baixa com lock otimista |
| `faturamento` | 8083 | `faturamento_db` | notas fiscais, itens, impressão, outbox de e-mail |
| `auth` | 8081 | `auth_db` | cadastro, verificação por e-mail, login, JWT |

`faturamento` chama `estoque` via HTTP (com timeout, retry e circuit breaker) na hora de imprimir
uma nota — nunca acessa o banco do estoque diretamente. `auth` e os demais serviços não têm
dependência entre si.

Stack: Go 1.25, Gin, `pgx/v5` com SQL puro (sem ORM), PostgreSQL 16, `golang-jwt/jwt/v5`,
`bcrypt`, `sony/gobreaker`, `go-playground/validator`, tudo em Docker. Frontend em Angular.

## Como rodar

```bash
docker compose up --build
```

Sobe os três serviços + Postgres + MailHog. Healthcheck garante que cada serviço só fica pronto
depois do banco responder, e migrations rodam sozinhas no boot de cada serviço (sem ferramenta
externa).

Depois de subir:

- Estoque: http://localhost:8082 (`GET /health`)
- Faturamento: http://localhost:8083 (`GET /health`)
- Auth: http://localhost:8081 (`GET /health`)
- MailHog (UI de e-mail): http://localhost:8025

Pra derrubar sem perder dados: `docker compose down`. Pra apagar tudo, incluindo o volume do
Postgres: `docker compose down -v`.

## Popular com dados de exemplo

```bash
./scripts/seed.sh
```

Precisa dos serviços já rodando (`docker compose up`). Cria produtos, uma nota fechada e uma
nota aberta pra já ter algo pra mostrar sem precisar cadastrar tudo na mão.

## Testar as rotas manualmente

`requests.http` na raiz — coleção com todas as rotas dos três serviços, compatível com a
extensão REST Client (VSCode) ou o HTTP Client do JetBrains. Cobre principais cenários de erro de cada endpoint.

## Rodar os testes

```bash
cd backend
go test ./...
```

Teste unitário na regra de baixa de estoque (saldo insuficiente, conflito de versão) — não depende de banco real (mocka a camada de acesso a dados).

## Principais endpoints

**Estoque** (`:8082`)
- `POST/GET /produtos`, `GET/PUT/DELETE /produtos/:id`
- `POST /estoque/baixas` — uso interno, chamado pelo faturamento na impressão
- `POST /produtos/sugestao` — opcional, sugere descrição + produtos similares via IA (ver abaixo)

**Faturamento** (`:8083`)
- `POST/GET /notas`, `GET/DELETE /notas/:id`
- `POST/PUT/DELETE /notas/:id/itens[/:itemId]`
- `POST /notas/:id/imprimir` — header `Idempotency-Key` obrigatório

**Auth** (`:8081`)
- `POST /auth/cadastro`, `GET /auth/verificar?token=...`, `POST /auth/login`
- `GET /auth/me` — protegido, header `Authorization: Bearer <token>`

Todo erro segue o mesmo formato em qualquer serviço:

```json
{ "code": "SALDO_INSUFICIENTE", "message": "...", "details": [], "trace_id": "..." }
```

## Feature opcional: sugestão de produto via IA

`POST /produtos/sugestao` (estoque) sugere descrição + produtos similares a partir só do
`codigo`, usando a API do Google Gemini. Precisa de uma chave gratuita:

1. Cria conta em [aistudio.google.com](https://aistudio.google.com) (sem cartão)
2. Gera uma chave em "Get API key"
3. Copia `.env.example` pra `.env` na raiz do projeto e cola a chave em `GEMINI_API_KEY`
4. `docker compose up --build`

Sem a chave configurada, o resto do sistema funciona normal — só esse endpoint específico
devolve `503 IA_INDISPONIVEL` em vez de derrubar o serviço.

## Resiliência

Impressão de nota fiscal chama o serviço de estoque com timeout de 3s, retry com backoff em
falha de rede/5xx, e circuit breaker (abre depois de 3 falhas seguidas). Uma goroutine ("reaper") roda a cada 30s e devolve pra `ABERTA` qualquer
nota travada em `PROCESSANDO` há mais de 2 minutos — protege contra o faturamento cair no meio
de uma impressão.

