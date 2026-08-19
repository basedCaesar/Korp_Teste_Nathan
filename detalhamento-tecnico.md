# Detalhamento Técnico

## Requisitos atendidos

| Requisito | Status | Onde |
|---|---|---|
| Obrigatório 1 — arquitetura de microsserviços (mín. 2) | ✅ | `estoque` + `faturamento` (+ `auth`, opcional) |
| Obrigatório 2 — tratamento de falhas, recuperação e feedback | ✅ | retry + circuit breaker + reaper; `ESTOQUE_INDISPONIVEL` 503; cenário testado em E2E derrubando o container `estoque` |
| Obrigatório 3 — conexão real com banco de dados | ✅ | Postgres, `pgx/v5`, um banco por serviço |
| Opcional (a) — tratamento de concorrência | ✅ | lock otimista na baixa de estoque, teste unitário 4/4 |
| Opcional (b) — uso de IA | ✅ | `POST /produtos/sugestao`, Google Gemini |
| Opcional (c) — idempotência | ✅ | tabela `idempotency_keys` + middleware, em `/imprimir` e `/estoque/baixas` |

## Arquitetura

Monorepo, um único `go.mod` para os três serviços backend. Frontend Angular em projeto separado.

| Serviço | Porta | Banco | Responsabilidade |
|---|---|---|---|
| estoque | 8082 | estoque_db | produtos, saldos, baixa com lock otimista, sugestão via IA |
| faturamento | 8083 | faturamento_db | notas fiscais, itens, impressão, outbox |
| auth | 8081 | auth_db | cadastro, verificação por e-mail, login, JWT HS256 |
| frontend | 4200 | — | Angular via nginx, config runtime via `/config.json` |

Cada domínio (`internal/estoque`, `internal/faturamento`, `internal/auth`) é um pacote Go
autocontido, em camadas: `handler` (transporte HTTP, bind/validação) → `service` (regra de
negócio) → `repository` (SQL puro, único lugar que fala com o banco). Código transversal
(envelope de erro, request-id, idempotência, JWT, pool de conexão, retry/circuit breaker) em
`internal/httpx`, `internal/db`, `internal/resilience`, usados pelos três domínios.

Serviços stateless — todo estado no Postgres, volume nomeado `pgdata`.

### Modelo de dados

| Banco | Tabela | Colunas principais |
|---|---|---|
| estoque_db | produtos | id, codigo (único por user_id), descricao, saldo, version, categoria, user_id |
| faturamento_db | notas | id, numero (sequence), status, user_id |
| faturamento_db | itens_nota | id, nota_id, produto_id, produto_codigo, produto_descricao, quantidade |
| faturamento_db | outbox | id, nota_id, destinatario, assunto, corpo, processado, tentativas |
| estoque_db / faturamento_db | idempotency_keys | chave (PK), status_code, corpo, created_at |
| auth_db | users | id, email (único), senha_hash, verificado, token_verificacao, token_verificacao_expira_em |

## Backend Go

### Frameworks utilizados

Gin (`gin-gonic/gin`) nos três serviços — `gin.New()` sem middlewares padrão, só
`gin.Recovery()` próprio via `httpx.Recovery()`. `log/slog` (stdlib) pra log estruturado em
JSON. Nenhum outro framework web.

### Gerenciamento de dependências no Golang

Go Modules (`go.mod`/`go.sum`), monorepo com módulo único `korp` pros três serviços. Sem
vendoring. Build multi-stage no Docker (`golang:1.22-alpine` compila com `CGO_ENABLED=0`,
`alpine:3.19` só copia o binário + `ca-certificates`).

Principais dependências:

| Pacote | Uso |
|---|---|
| `pgx/v5` | driver Postgres, SQL puro, sem ORM |
| `gin-gonic/gin` | HTTP |
| `go-playground/validator/v10` | validação de payload |
| `golang-jwt/jwt/v5` | JWT HS256 |
| `golang.org/x/crypto/bcrypt` | hash de senha |
| `sony/gobreaker` | circuit breaker |

### Tratamento de erros e exceções

Envelope único em toda resposta de erro dos três serviços: `{code, message, details, trace_id}`.

Erros de domínio são sentinelas Go (`errors.New`), classificadas na camada HTTP via
`errors.Is`/`errors.As` — o handler nunca monta o JSON de erro na mão. O repository nunca
devolve erro de infra cru: mapeia `pgx.ErrNoRows` para sentinela de "não encontrado" e violação
de unique constraint do Postgres (`pgconn.PgError.Code == "23505"`) para sentinela de
"duplicado". Erro de binding/validação tem tradutor próprio (`httpx.RespondValidationError`),
lista cada campo inválido em `details`. Panic é recuperado por `httpx.Recovery()` e cai no mesmo
envelope (`ERRO_INTERNO`, 500).

`trace_id` é o `X-Request-Id` da requisição (usa o header recebido ou gera um novo, propagado
entre serviços via header, devolvido no envelope).

Exceção deliberada: `GET /health` e `GET /health/dependencias` sempre respondem `200` — status
de cada dependência vai no corpo, não no código HTTP (evita disparar interceptor de erro do
frontend a cada poll).

Códigos por serviço:

**Comuns:** `VALIDACAO` 400 · `ID_INVALIDO` 400 · `TOKEN_AUSENTE`/`TOKEN_INVALIDO` 401 ·
`ERRO_INTERNO` 500

**estoque:** `PRODUTO_NAO_ENCONTRADO` 404 · `CODIGO_DUPLICADO` 409 · `SALDO_INSUFICIENTE` 409 ·
`CONFLITO_VERSAO` 409 · `IA_INDISPONIVEL` 503

**faturamento:** `NOTA_NAO_ENCONTRADA`/`ITEM_NAO_ENCONTRADO` 404 · `NOTA_NAO_ABERTA` 409 ·
`SALDO_INSUFICIENTE` 409 · `ESTOQUE_INDISPONIVEL` 503 · `PRODUTO_INVALIDO` 400 ·
`IDEMPOTENCY_KEY_OBRIGATORIA` 400

**auth:** `EMAIL_JA_CADASTRADO` 409 · `CREDENCIAIS_INVALIDAS` 401 · `EMAIL_NAO_VERIFICADO` 403 ·
`TOKEN_INVALIDO`/`TOKEN_EXPIRADO` 400 (token de verificação de e-mail)

### Regras de negócio

**Baixa de estoque** (lock otimista): `UPDATE produtos SET saldo = saldo - $2, version =
version + 1 WHERE id = $1 AND version = $3 AND saldo >= $2`. Zero linhas afetadas → relê;
`saldo < quantidade` → `SALDO_INSUFICIENTE`; versão mudou → retenta até 3x, senão
`CONFLITO_VERSAO`. Itens ordenados por `produto_id` antes de aplicar, evita deadlock entre
baixas concorrentes.

**Impressão** (`POST /notas/{id}/imprimir`, `Idempotency-Key` obrigatório): transição atômica
`ABERTA→PROCESSANDO` trava contra duplo clique/corrida; chama `POST /estoque/baixas` (timeout
3s, retry com backoff só em 5xx/timeout, circuit breaker via `sony/gobreaker`, abre com 3 falhas
consecutivas); sucesso fecha a nota e insere na `outbox` na mesma transação; falha volta pra
`ABERTA` e devolve o erro real (`SALDO_INSUFICIENTE` 409 ou `ESTOQUE_INDISPONIVEL` 503).

**Reaper**: goroutine com ticker de 30s, devolve para `ABERTA` notas em `PROCESSANDO` há mais de
2 minutos — cobre queda do processo faturamento no meio da impressão.

**Idempotência**: tabela `idempotency_keys` (chave, status, corpo, timestamp) + middleware
genérico em `internal/httpx`, aplicado em `POST /notas/{id}/imprimir` e `POST /estoque/baixas`.

**Outbox**: insert na mesma transação que fecha a nota; consumidor assíncrono (ticker de 5s)
envia e-mail via SMTP, marca processado ou incrementa tentativas (limite 5).

**Auth**: cadastro com bcrypt, token de verificação por e-mail (24h, uso único), login gera JWT
HS256 (claims `user_id`/`email`, expira 24h). `httpx.JWTAuth(secret)` genérico protege os
grupos `/produtos` e `/notas` em estoque e faturamento (mesmo `JWT_SECRET` nos três serviços) —
`/estoque/baixas` fica fora, é chamada interna serviço-a-serviço.

**Multi-tenant**: produto e nota têm `user_id`, toda query filtra por dono. Sem FK real entre
serviços (bancos Postgres separados) — a garantia é o JWT assinado com segredo compartilhado.
Acesso a recurso de outro dono devolve 404 (não 403), reaproveitando as sentinelas de "não
encontrado" já existentes. `POST /notas/{id}/itens` valida o produto referenciado contra o
estoque de verdade (`GET /produtos/{id}` repassando o `Authorization` do usuário), evitando
referenciar `produto_id` de outra conta.

**IA (sugestão de produto)**: `POST /produtos/sugestao`, Google Gemini via HTTP puro (sem SDK),
modelo configurável por env var. Endpoint isolado, não persiste nada. Vê o catálogo global
mesmo com multi-tenant ativo — é a única exceção ao isolamento por usuário. Campo `categoria`
opcional em produto pré-filtra o catálogo antes de montar o prompt, quando informado.

**CORS**: middleware genérico (`internal/httpx/cors.go`), origem configurável via
`CORS_ALLOWED_ORIGIN` (default `*` — token vai em header `Authorization`, não em cookie, sem
risco de CSRF). Preflight `OPTIONS` respondido direto com `204`.

## Frontend Angular

### Ciclos de vida utilizados

`ngOnInit` nas telas que carregam dado inicial (produtos, notas, detalhe da nota). Sem
`ngOnDestroy` manual na maior parte do app — zoneless, chamadas HTTP via `HttpClient` completam
sozinhas, não são streams infinitas. Exceção: `nota-detalhe.ts` implementa `OnDestroy` de
verdade pro polling de status do estoque (`StatusSistemaService`, singleton `providedIn:
'root'`) — liga em `ngOnInit`, desliga em `ngOnDestroy`, par explícito `iniciar()`/`parar()`.

### RxJS

- `debounceTime` + `distinctUntilChanged` + `switchMap` — sugestão via IA ao digitar
  código/categoria do produto (600ms de debounce, cancela chamada anterior em voo).
- `catchError` — interceptor de erro global, extrai `{code, message, details, trace_id}` e
  mostra em snackbar.
- `finalize` — liga/desliga spinners (botão Imprimir, dialogs, listas) independente de
  sucesso/erro.
- `firstValueFrom` — converte bootstrap de config (`GET /config.json`) em Promise para
  `provideAppInitializer`.
- `interval` + `startWith` + `switchMap` — polling de saúde do estoque a cada 5s, erro de rede
  vira `catchError(() => of(false))` pra não matar a stream.

### Outras bibliotecas

Nenhuma além do que Angular/Material trazem — `rxjs` e `@angular/cdk` (transitiva do Material,
overlay de dialog/autocomplete). Sem `date-fns`/`moment` (`DatePipe` resolve), sem cliente HTTP
alternativo.

### Bibliotecas de componentes visuais

Angular Material — tema Material 3 customizado (`mat.theme()`, paleta violet/cyan, densidade
-1). Componentes: `MatTable`, `MatDialog`, `MatAutocomplete`, `MatSnackBar`,
`MatFormField`/`MatInput`, `MatToolbar`, `MatButton`, `MatIcon`, `MatProgressSpinner`,
`MatCard`. Status da nota (`ABERTA`/`PROCESSANDO`/`FECHADA`) é `<span>` com classe própria, não
`MatChip` — precisa de cor sólida por status sem sobrescrever CSS interno do componente.

### Config em runtime

`provideAppInitializer` busca `GET /config.json` antes do app renderizar — em dev vem estático
de `public/config.json`; em produção, `docker-entrypoint.sh` gera o arquivo a partir de env vars
(`ESTOQUE_URL`, `FATURAMENTO_URL`, `AUTH_URL`) no boot do container nginx. Sem URL hardcoded em
build-time.

### Autenticação no frontend

JWT no `localStorage`, payload decodificado (`atob`) só para exibição (`email`/`user_id` na
toolbar) — nunca usado pra autorizar nada no cliente. `authTokenInterceptor` (funcional) anexa
`Authorization: Bearer` em toda requisição. `authGuard` (`CanActivateFn`) bloqueia
`/produtos`, `/notas`, `/notas/:id` sem login. Isolamento de dado é responsabilidade do
backend — o frontend só exibe o que a API já devolve filtrado.

## Testes

| Camada | Ferramenta | Resultado |
|---|---|---|
| Backend, unitário | `go test` | 4/4 — regra de baixa de estoque (saldo insuficiente, conflito de versão) |
| Frontend, unitário | Vitest (`ng test`) | 45/45 — services, interceptors, `AuthService`, componentes principais |
| Integração (front+back) | Playwright, E2E | 13/14 — 1 skip dinâmico esperado (flakiness da IA), inclui cenário obrigatório de falha/recuperação do estoque |

Backend fora da baixa de estoque validado manualmente contra os containers reais (CRUD,
concorrência real, circuit breaker cronometrado, reaper com crash forçado, outbox ponta a
ponta) — decisão de escopo, não teste automatizado nesses casos.

## Melhorias futuras

- Refresh token — JWT de 24h sem revogação hoje.
- Rate limiting em `/auth/login` e `/produtos/sugestao`.
- Paginação em `GET /produtos`/`GET /notas`.
- `user_id` cross-serviço confia só no JWT — só item de nota reconfirma dono de verdade contra o estoque.
- Sem recuperação de senha no `auth`, sem métrica além do log estruturado.
