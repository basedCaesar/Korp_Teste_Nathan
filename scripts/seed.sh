#!/usr/bin/env bash
set -euo pipefail

ESTOQUE_URL="${ESTOQUE_URL:-http://localhost:8082}"
FATURAMENTO_URL="${FATURAMENTO_URL:-http://localhost:8083}"
AUTH_URL="${AUTH_URL:-http://localhost:8081}"
MAILHOG_URL="${MAILHOG_URL:-http://localhost:8025}"
SEED_EMAIL="${SEED_EMAIL:-seed@korp.local}"
SEED_SENHA="${SEED_SENHA:-seed12345}"

extrair_id() {
  grep -o '"id":[0-9]*' | head -1 | grep -o '[0-9]*'
}

echo "Aguardando servicos ficarem saudaveis..."
for url in "$ESTOQUE_URL/health" "$FATURAMENTO_URL/health" "$AUTH_URL/health"; do
  for _ in $(seq 1 30); do
    if curl -sf "$url" > /dev/null; then
      break
    fi
    sleep 1
  done
done

echo "Autenticando usuario seed ($SEED_EMAIL)..."
status_cadastro=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$AUTH_URL/auth/cadastro" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$SEED_EMAIL\",\"senha\":\"$SEED_SENHA\"}")

if [ "$status_cadastro" = "201" ]; then
  echo "  conta nova, verificando e-mail via MailHog..."
  token=""
  for _ in $(seq 1 15); do
    token=$(curl -s "$MAILHOG_URL/api/v2/search?kind=to&query=$SEED_EMAIL" \
      | grep -o 'token=[a-f0-9]*' | head -1 | cut -d= -f2)
    [ -n "$token" ] && break
    sleep 1
  done
  if [ -z "$token" ]; then
    echo "  nao achei o token de verificacao no MailHog ($MAILHOG_URL) a tempo." >&2
    exit 1
  fi
  curl -s "$AUTH_URL/auth/verificar?token=$token" > /dev/null
fi

JWT=$(curl -s -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$SEED_EMAIL\",\"senha\":\"$SEED_SENHA\"}" \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$JWT" ]; then
  echo "  login do usuario seed falhou." >&2
  exit 1
fi

auth=(-H "Authorization: Bearer $JWT")

ja_semeado=$(curl -s "$ESTOQUE_URL/produtos" "${auth[@]}" | grep -o '"codigo":"SEED-001"' || true)
if [ -n "$ja_semeado" ]; then
  echo "Usuario seed ($SEED_EMAIL) ja tem os dados de exemplo. Nada a fazer."
  exit 0
fi

echo "Criando produtos..."
P1=$(curl -s -X POST "$ESTOQUE_URL/produtos" "${auth[@]}" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-001","descricao":"Teclado mecanico","saldo":50}' | extrair_id)
P2=$(curl -s -X POST "$ESTOQUE_URL/produtos" "${auth[@]}" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-002","descricao":"Mouse sem fio","saldo":30}' | extrair_id)
P3=$(curl -s -X POST "$ESTOQUE_URL/produtos" "${auth[@]}" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-003","descricao":"Monitor 27 polegadas","saldo":15}' | extrair_id)
echo "  produtos: $P1 $P2 $P3"

echo "Criando nota fechada (com impressao)..."
N1=$(curl -s -X POST "$FATURAMENTO_URL/notas" "${auth[@]}" | extrair_id)
curl -s -X POST "$FATURAMENTO_URL/notas/$N1/itens" "${auth[@]}" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P1,\"produto_codigo\":\"SEED-001\",\"produto_descricao\":\"Teclado mecanico\",\"quantidade\":2}" > /dev/null
curl -s -X POST "$FATURAMENTO_URL/notas/$N1/itens" "${auth[@]}" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P2,\"produto_codigo\":\"SEED-002\",\"produto_descricao\":\"Mouse sem fio\",\"quantidade\":5}" > /dev/null
curl -s -o /dev/null -X POST "$FATURAMENTO_URL/notas/$N1/imprimir" "${auth[@]}" -H "Idempotency-Key: seed-$N1"
echo "  nota $N1 fechada"

echo "Criando nota aberta (sem imprimir)..."
N2=$(curl -s -X POST "$FATURAMENTO_URL/notas" "${auth[@]}" | extrair_id)
curl -s -X POST "$FATURAMENTO_URL/notas/$N2/itens" "${auth[@]}" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P3,\"produto_codigo\":\"SEED-003\",\"produto_descricao\":\"Monitor 27 polegadas\",\"quantidade\":1}" > /dev/null
echo "  nota $N2 aberta"

echo ""
echo "Seed concluido:"
echo "  usuario: $SEED_EMAIL / $SEED_SENHA"
echo "  produtos: $P1, $P2, $P3"
echo "  nota fechada: $N1"
echo "  nota aberta: $N2"
