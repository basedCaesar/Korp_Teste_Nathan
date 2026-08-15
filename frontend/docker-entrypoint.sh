#!/bin/sh
set -eu

cat > /usr/share/nginx/html/config.json <<EOF
{
  "estoqueUrl": "${ESTOQUE_URL:-http://localhost:8082}",
  "faturamentoUrl": "${FATURAMENTO_URL:-http://localhost:8083}",
  "authUrl": "${AUTH_URL:-http://localhost:8081}"
}
EOF

exec "$@"
