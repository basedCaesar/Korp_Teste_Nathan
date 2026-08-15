-- Roda uma vez, no primeiro start do container (volume pgdata vazio).
-- Cria um database por serviço; migrations de tabela ficam em internal/db,
-- rodadas no boot de cada serviço.

CREATE DATABASE estoque_db;
CREATE DATABASE faturamento_db;
CREATE DATABASE auth_db;
