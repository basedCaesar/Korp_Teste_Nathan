import { expect, test } from '@playwright/test';

import { criarContaVerificada, loginViaUI } from './helpers/auth';
import { criarProdutoViaAPI } from './helpers/produtos';

test('cria nota, adiciona/edita/remove item e exclui a nota pela UI', async ({
  page,
  request,
}) => {
  const conta = await criarContaVerificada(request, 'e2e-notas');
  const produto = await criarProdutoViaAPI(request, conta.token, { saldo: 20 });
  await loginViaUI(page, conta);

  await page.goto('/notas');
  await page.getByRole('button', { name: 'Nova nota' }).click();
  await page.waitForURL(/\/notas\/\d+/);
  await expect(page.getByText('ABERTA')).toBeVisible();

  await page.getByRole('button', { name: 'Adicionar produto' }).click();
  await page.getByLabel('Produto', { exact: true }).fill(produto.codigo);
  await page.getByRole('option', { name: new RegExp(produto.codigo) }).click();
  await page.getByLabel('Quantidade', { exact: true }).fill('3');
  await page.locator('mat-dialog-container').getByRole('button', { name: 'Salvar' }).click();

  const linhaItem = page.locator('tr', { hasText: produto.codigo });
  await expect(linhaItem.locator('td').nth(1)).toHaveText('3');

  await linhaItem.getByRole('button', { name: 'Editar quantidade' }).click();
  await page.getByLabel('Quantidade', { exact: true }).fill('7');
  await page.locator('mat-dialog-container').getByRole('button', { name: 'Salvar' }).click();
  await expect(linhaItem.locator('td').nth(1)).toHaveText('7');

  page.once('dialog', (dialog) => dialog.accept());
  await linhaItem.getByRole('button', { name: 'Remover item' }).click();
  await expect(page.locator('tr', { hasText: produto.codigo })).toHaveCount(0);
  await expect(page.getByText('Nenhum item nesta nota ainda.')).toBeVisible();

  page.once('dialog', (dialog) => dialog.accept());
  await page.getByRole('button', { name: 'Excluir nota' }).click();
  await page.waitForURL('**/notas');
});
