import { APIRequestContext, expect } from '@playwright/test';

import { ESTOQUE_URL } from './urls';

export async function criarProdutoViaAPI(
  request: APIRequestContext,
  token: string,
  overrides: Partial<{ codigo: string; descricao: string; saldo: number }> = {},
) {
  const codigo = overrides.codigo ?? `E2E-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
  const resp = await request.post(`${ESTOQUE_URL}/produtos`, {
    headers: { Authorization: `Bearer ${token}` },
    data: {
      codigo,
      descricao: overrides.descricao ?? 'Produto de teste E2E',
      saldo: overrides.saldo ?? 10,
    },
  });
  expect(resp.status(), await resp.text()).toBe(201);
  return resp.json() as Promise<{ id: number; codigo: string; descricao: string; saldo: number }>;
}
