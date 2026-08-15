#!/usr/bin/env bash
set -euo pipefail

ESTOQUE_URL="${ESTOQUE_URL:-http://localhost:8082}"
FATURAMENTO_URL="${FATURAMENTO_URL:-http://localhost:8083}"

extrair_id() {
  grep -o '"id":[0-9]*' | head -1 | grep -o '[0-9]*'
}

echo "Aguardando servicos ficarem saudaveis..."
for url in "$ESTOQUE_URL/health" "$FATURAMENTO_URL/health"; do
  for _ in $(seq 1 30); do
    if curl -sf "$url" > /dev/null; then
      break
    fi
    sleep 1
  done
done

echo "Criando produtos..."
P1=$(curl -s -X POST "$ESTOQUE_URL/produtos" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-001","descricao":"Teclado mecanico","saldo":50}' | extrair_id)
P2=$(curl -s -X POST "$ESTOQUE_URL/produtos" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-002","descricao":"Mouse sem fio","saldo":30}' | extrair_id)
P3=$(curl -s -X POST "$ESTOQUE_URL/produtos" -H "Content-Type: application/json" \
  -d '{"codigo":"SEED-003","descricao":"Monitor 27 polegadas","saldo":15}' | extrair_id)
echo "  produtos: $P1 $P2 $P3"

echo "Criando nota fechada (com impressao)..."
N1=$(curl -s -X POST "$FATURAMENTO_URL/notas" | extrair_id)
curl -s -X POST "$FATURAMENTO_URL/notas/$N1/itens" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P1,\"produto_codigo\":\"SEED-001\",\"produto_descricao\":\"Teclado mecanico\",\"quantidade\":2}" > /dev/null
curl -s -X POST "$FATURAMENTO_URL/notas/$N1/itens" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P2,\"produto_codigo\":\"SEED-002\",\"produto_descricao\":\"Mouse sem fio\",\"quantidade\":5}" > /dev/null
curl -s -o /dev/null -X POST "$FATURAMENTO_URL/notas/$N1/imprimir" -H "Idempotency-Key: seed-$N1"
echo "  nota $N1 fechada"

echo "Criando nota aberta (sem imprimir)..."
N2=$(curl -s -X POST "$FATURAMENTO_URL/notas" | extrair_id)
curl -s -X POST "$FATURAMENTO_URL/notas/$N2/itens" -H "Content-Type: application/json" \
  -d "{\"produto_id\":$P3,\"produto_codigo\":\"SEED-003\",\"produto_descricao\":\"Monitor 27 polegadas\",\"quantidade\":1}" > /dev/null
echo "  nota $N2 aberta"

echo ""
echo "Seed concluido:"
echo "  produtos: $P1, $P2, $P3"
echo "  nota fechada: $N1"
echo "  nota aberta: $N2"
