# Korp Teste — Sistema de Emissão de Notas Fiscais

Teste técnico: sistema de emissão de notas fiscais em arquitetura de microsserviços, backend em
Go e frontend em Angular. Cada usuário só vê e mexe nos próprios produtos e notas — login é
obrigatório pra usar o sistema.

## Arquitetura

Três serviços de backend independentes, cada um com seu próprio banco Postgres, mais um SMTP
fake (MailHog) pra capturar e-mails sem depender de conta real, mais o frontend:

| Serviço | Porta | Database | Responsabilidade |
|---|---|---|---|
| `estoque` | 8082 | `estoque_db` | produtos (por usuário), saldos, baixa com lock otimista |
| `faturamento` | 8083 | `faturamento_db` | notas fiscais (por usuário), itens, impressão, outbox de e-mail |
| `auth` | 8081 | `auth_db` | cadastro, verificação por e-mail, login, JWT |
| `frontend` | 4200 | — | Angular, servido via nginx |

`faturamento` chama `estoque` via HTTP (com timeout, retry e circuit breaker) na hora de imprimir
uma nota — nunca acessa o banco do estoque diretamente. `estoque` e `faturamento` validam o JWT
emitido pelo `auth` (mesmo `JWT_SECRET` nos três) pra saber de quem é cada produto/nota — nenhum
dos dois chama o `auth` de volta, o token já vem assinado e autocontido.

**Isolamento por usuário:** produtos e notas pertencem a quem os criou (`user_id` do token).
Um usuário nunca vê, edita nem referencia produto/nota de outro — inclusive por chamada direta
na API (adicionar item numa nota com `produto_id` de outro usuário é rejeitado,
`400 PRODUTO_INVALIDO`). A única exceção proposital é a sugestão via IA
(`POST /produtos/sugestao`): ela enxerga o catálogo de todo mundo pra poder sugerir com base em
produtos parecidos, mesmo que não sejam do usuário logado.

Stack: Go 1.25, Gin, `pgx/v5` com SQL puro (sem ORM), PostgreSQL 16, `golang-jwt/jwt/v5`,
`bcrypt`, `sony/gobreaker`, `go-playground/validator`, tudo em Docker. Frontend em Angular
(standalone components, Angular Material, RxJS, Signals) — detalhes técnicos em `detalhamento.md`.

## Como rodar

```bash
docker compose up --build
```

Sobe os três serviços de backend + frontend + Postgres + MailHog. Healthcheck garante que cada
serviço só fica pronto depois do banco responder, e migrations rodam sozinhas no boot de cada
serviço (sem ferramenta externa).

Depois de subir:

- **Frontend: http://localhost:4200** — abre direto no login, precisa cadastrar uma conta antes
  de usar (link "Cadastrar" na própria tela)
- Estoque: http://localhost:8082 (`GET /health`)
- Faturamento: http://localhost:8083 (`GET /health`)
- Auth: http://localhost:8081 (`GET /health`)
- MailHog (UI de e-mail, usado pra pegar o link de verificação de cadastro): http://localhost:8025

Pra derrubar sem perder dados: `docker compose down`. Pra apagar tudo, incluindo o volume do
Postgres (recomendado antes de gravar o vídeo, pra começar com base limpa): `docker compose down -v`.

## Popular com dados de exemplo

```bash
./scripts/seed.sh
```

Precisa dos serviços já rodando (`docker compose up`). Cadastra e verifica um usuário de teste
(`seed@korp.local` / `seed12345`, verificação lida direto da API do MailHog — não precisa clicar
em nada), e só então cria produtos, uma nota fechada e uma nota aberta **pertencentes a esse
usuário**. Idempotente: rodar de novo detecta que já foi semeado e não duplica nada — dá pra usar
como smoke test rápido depois de qualquer mudança. Pra ver os dados no frontend, loga com esse
mesmo usuário.

## Testar as rotas manualmente

`requests.http` na raiz — coleção com todas as rotas dos três serviços de backend, compatível
com a extensão REST Client (VSCode) ou o HTTP Client do JetBrains. Cobre principais cenários de
erro de cada endpoint. Roda a seção **AUTH** primeiro (cadastro → verificar com o token do
MailHog → login) — as seções de produtos/notas reusam o token do login automaticamente
(`{{login.response.body.$.token}}`, suportado nativamente pelo REST Client).

## Rodar os testes

```bash
cd backend
go test ./...
```

Teste unitário na regra de baixa de estoque (saldo insuficiente, conflito de versão) — não depende de banco real (mocka a camada de acesso a dados).

```bash
cd frontend
npm test
```

45 testes unitários (Vitest) — services, interceptors, `AuthService`, componentes principais.
**Precisa de Node 22+** (exigência do builder novo do Angular, não do projeto em si). Sem
instalar Node 22 localmente, roda isolado num container:

```bash
docker run --rm -v "$(pwd)":/app -w /app node:22-alpine sh -c "npm ci && npx ng test --watch=false"
```

## Testes end-to-end (frontend + backend juntos)

```bash
docker compose up -d   # precisa da stack real de pé
cd frontend
npm ci
npm run e2e
```

Só precisa de Node **20+** (o script roda o Playwright direto, não passa pelo Angular CLI) —
não precisa da versão 22 exigida pelo teste unitário acima.

Playwright rodando contra a aplicação real (não mock nenhum) — cobre produtos, notas, itens,
impressão (sucesso e o cenário obrigatório de falha/recuperação, derrubando e subindo o
container `estoque` de verdade a partir do próprio teste), auth e isolamento entre usuários.
14 specs (13 passam, 1 pula se a IA — Gemini — estiver indisponível no momento, não é bug).

## Principais endpoints

**Estoque** (`:8082`) — tudo abaixo exige `Authorization: Bearer <token>`, exceto `/estoque/baixas`
- `POST/GET /produtos`, `GET/PUT/DELETE /produtos/:id` — só do usuário do token
- `POST /estoque/baixas` — uso interno, chamado pelo faturamento na impressão, sem token (não é
  um usuário fazendo a chamada, é serviço-a-serviço)
- `POST /produtos/sugestao` — opcional, sugere descrição + produtos similares via IA a partir do
  catálogo de **todos** os usuários (ver abaixo)

**Faturamento** (`:8083`) — tudo abaixo exige `Authorization: Bearer <token>`
- `POST/GET /notas`, `GET/DELETE /notas/:id` — só do usuário do token
- `POST/PUT/DELETE /notas/:id/itens[/:itemId]` — `produto_id` precisa ser de um produto do
  mesmo usuário, senão `400 PRODUTO_INVALIDO`
- `POST /notas/:id/imprimir` — header `Idempotency-Key` obrigatório. Sucesso devolve `200` com
  corpo vazio (não a nota); pra ver o status `FECHADA` e o resultado, faz `GET /notas/:id` em
  seguida.
- `GET /health/dependencias` — sem auth, sempre `200`. Corpo `{"dependencias": {"estoque":
  true|false}}`, usado pelo frontend pra desabilitar o botão Imprimir e avisar o usuário
  proativamente quando o `estoque` está fora do ar (polling a cada 5s enquanto uma nota está
  aberta na tela).

**Auth** (`:8081`)
- `POST /auth/cadastro`, `GET /auth/verificar?token=...`, `POST /auth/login`
- `GET /auth/me` — protegido, header `Authorization: Bearer <token>`

Todo erro segue o mesmo formato em qualquer serviço:

```json
{ "code": "SALDO_INSUFICIENTE", "message": "...", "details": [], "trace_id": "..." }
```

Produto/nota de outro usuário (ou inexistente) sempre devolve `404`, nunca `403` — não confirma
pra quem tenta acessar que aquele id existe e é de outra conta.

## Feature opcional: sugestão de produto via IA

`POST /produtos/sugestao` (estoque) sugere descrição + produtos similares a partir do `codigo`
e, opcionalmente, `categoria` — usando a API do Google Gemini. Produto tem um campo `categoria`
livre (texto, opcional); informar categoria na sugestão pré-filtra o catálogo só pra produtos
daquela categoria antes de perguntar pra IA, deixando a sugestão mais precisa (sem categoria,
a IA tem que adivinhar a categoria só pelo padrão do código). No frontend, o campo sugere as
categorias que o próprio usuário já usou (autocomplete), mas aceita digitar uma nova. Precisa
de uma chave gratuita pra usar o endpoint:

1. Cria conta em [aistudio.google.com](https://aistudio.google.com) (sem cartão)
2. Gera uma chave em "Get API key"
3. Copia `.env.example` pra `.env` na raiz do projeto e cola a chave em `GEMINI_API_KEY`
4. `docker compose up --build`

Sem a chave configurada, o resto do sistema funciona normal — só esse endpoint específico
devolve `503 IA_INDISPONIVEL` em vez de derrubar o serviço.

**Troubleshooting:** se o endpoint devolver `503 IA_INDISPONIVEL` mesmo com a chave certa, três
causas possíveis, cada uma com sintoma diferente:

1. **Modelo sobrecarregado do lado do Google** — resposta rápida, `503 UNAVAILABLE`, mensagem
   "This model is currently experiencing high demand". Tenta de novo em alguns segundos.
2. **Rede do host** — trava ~15s (o timeout do `http.Client`) toda vez antes de falhar. Em
   Docker Desktop/WSL2 o container pode tentar IPv6 pra `generativelanguage.googleapis.com` e
   não receber resposta nenhuma; por isso `ia.go` força IPv4 (`DialContext` com rede `tcp4`).
   Se travar 15s mesmo assim, é rede do host, não a chave.
3. **Cota diária esgotada** — resposta rápida, `429 RESOURCE_EXHAUSTED` do lado do Gemini
   (repassado como `503` pelo backend). Free tier tem limite de 20 requisições/dia **por
   projeto Google**, não por chave — gerar uma chave nova na mesma conta não resolve, as duas
   compartilham a cota. Só reseta sozinho (diário) ou usando outro projeto Google
   (aistudio.google.com → criar novo projeto → gerar chave nesse projeto).

## Resiliência

Impressão de nota fiscal chama o serviço de estoque com timeout de 3s, retry com backoff em
falha de rede/5xx, e circuit breaker (abre depois de 3 falhas seguidas). Uma goroutine ("reaper") roda a cada 30s e devolve pra `ABERTA` qualquer
nota travada em `PROCESSANDO` há mais de 2 minutos — protege contra o faturamento cair no meio
de uma impressão.

