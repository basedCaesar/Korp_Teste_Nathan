import { expect, test } from '@playwright/test';

import { criarContaVerificada, loginViaUI } from './helpers/auth';
import { criarProdutoViaAPI } from './helpers/produtos';
import { ESTOQUE_URL, FATURAMENTO_URL } from './helpers/urls';

test('cada usuario ve so os proprios produtos na UI', async ({ page, request }) => {
  const contaA = await criarContaVerificada(request, 'e2e-mt-a');
  const contaB = await criarContaVerificada(request, 'e2e-mt-b');
  const produtoA = await criarProdutoViaAPI(request, contaA.token, { descricao: 'Produto de A' });
  const produtoB = await criarProdutoViaAPI(request, contaB.token, { descricao: 'Produto de B' });

  await loginViaUI(page, contaA);
  await expect(page.locator('tr', { hasText: produtoA.codigo })).toBeVisible();
  await expect(page.locator('tr', { hasText: produtoB.codigo })).toHaveCount(0);

  await page.getByRole('button', { name: 'Sair' }).click();
  await page.waitForURL('**/login');

  await loginViaUI(page, contaB);
  await expect(page.locator('tr', { hasText: produtoB.codigo })).toBeVisible();
  await expect(page.locator('tr', { hasText: produtoA.codigo })).toHaveCount(0);
});

test('API rejeita item de nota referenciando produto de outro usuario', async ({ request }) => {
  const contaA = await criarContaVerificada(request, 'e2e-mt-api-a');
  const contaB = await criarContaVerificada(request, 'e2e-mt-api-b');
  const produtoA = await criarProdutoViaAPI(request, contaA.token);

  const notaB = await request.post(`${FATURAMENTO_URL}/notas`, {
    headers: { Authorization: `Bearer ${contaB.token}` },
  });
  const { id: notaId } = await notaB.json();

  const tentativa = await request.post(`${FATURAMENTO_URL}/notas/${notaId}/itens`, {
    headers: { Authorization: `Bearer ${contaB.token}` },
    data: {
      produto_id: produtoA.id,
      produto_codigo: produtoA.codigo,
      produto_descricao: produtoA.descricao,
      quantidade: 1,
    },
  });

  expect(tentativa.status()).toBe(400);
  expect((await tentativa.json()).code).toBe('PRODUTO_INVALIDO');
});

test('produto de outro usuario devolve 404 direto por id', async ({ request }) => {
  const contaA = await criarContaVerificada(request, 'e2e-mt-404-a');
  const contaB = await criarContaVerificada(request, 'e2e-mt-404-b');
  const produtoA = await criarProdutoViaAPI(request, contaA.token);

  const resp = await request.get(`${ESTOQUE_URL}/produtos/${produtoA.id}`, {
    headers: { Authorization: `Bearer ${contaB.token}` },
  });

  expect(resp.status()).toBe(404);
});
