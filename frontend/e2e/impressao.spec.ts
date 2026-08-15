import { expect, test } from '@playwright/test';

import { criarContaVerificada, loginViaUI } from './helpers/auth';
import { esperarEstoqueSaudavel, iniciarEstoque, pararEstoque } from './helpers/docker';
import { criarProdutoViaAPI } from './helpers/produtos';
import { ESTOQUE_URL } from './helpers/urls';

async function criarNotaComItemViaUI(page: import('@playwright/test').Page, codigoProduto: string) {
  await page.goto('/notas');
  await page.getByRole('button', { name: 'Nova nota' }).click();
  await page.waitForURL(/\/notas\/\d+/);

  await page.getByRole('button', { name: 'Adicionar produto' }).click();
  await page.getByLabel('Produto', { exact: true }).fill(codigoProduto);
  await page.getByRole('option', { name: new RegExp(codigoProduto) }).click();
  await page.getByLabel('Quantidade', { exact: true }).fill('2');
  await page.locator('mat-dialog-container').getByRole('button', { name: 'Salvar' }).click();
  await expect(page.locator('tr', { hasText: codigoProduto })).toBeVisible();
}

test('imprime com sucesso: status fecha e saldo do produto baixa de verdade', async ({
  page,
  request,
}) => {
  const conta = await criarContaVerificada(request, 'e2e-imprimir-ok');
  const produto = await criarProdutoViaAPI(request, conta.token, { saldo: 10 });
  await loginViaUI(page, conta);

  await criarNotaComItemViaUI(page, produto.codigo);

  await page.getByRole('button', { name: 'Imprimir' }).click();
  await expect(page.getByText('Nota fiscal impressa com sucesso.')).toBeVisible({
    timeout: 15_000,
  });
  await expect(page.getByText('FECHADA')).toBeVisible();
  await expect(page.getByRole('button', { name: 'Imprimir' })).toHaveCount(0);

  const produtoAtualizado = await request.get(`${ESTOQUE_URL}/produtos/${produto.id}`, {
    headers: { Authorization: `Bearer ${conta.token}` },
  });
  expect((await produtoAtualizado.json()).saldo).toBe(8);
});

test('falha do estoque: erro tratado, nota continua aberta, recupera sozinha depois', async ({
  page,
  request,
}) => {
  test.setTimeout(90_000);

  const conta = await criarContaVerificada(request, 'e2e-imprimir-falha');
  const produto = await criarProdutoViaAPI(request, conta.token, { saldo: 10 });
  await loginViaUI(page, conta);

  await criarNotaComItemViaUI(page, produto.codigo);

  pararEstoque();
  try {
    await page.getByRole('button', { name: 'Imprimir' }).click();
    await expect(page.getByText(/estoque indisponivel/i)).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText('ABERTA')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Imprimir' })).toBeVisible();
  } finally {
    iniciarEstoque();
    await esperarEstoqueSaudavel();
  }

  await page.getByRole('button', { name: 'Imprimir' }).click();
  await expect(page.getByText('Nota fiscal impressa com sucesso.')).toBeVisible({
    timeout: 15_000,
  });
  await expect(page.getByText('FECHADA')).toBeVisible();
});
