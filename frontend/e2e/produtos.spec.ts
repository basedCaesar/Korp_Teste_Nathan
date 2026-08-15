import { expect, test } from '@playwright/test';

import { criarContaVerificada, loginViaUI } from './helpers/auth';

test.describe('produtos', () => {
  test('cria, edita e exclui um produto pela UI', async ({ page, request }) => {
    const conta = await criarContaVerificada(request, 'e2e-produtos');
    await loginViaUI(page, conta);

    const codigo = `E2E-UI-${Date.now()}`;

    await page.getByRole('button', { name: 'Novo produto' }).click();
    await page.getByLabel('Código').fill(codigo);
    await page.getByLabel('Descrição').fill('Produto criado no E2E');
    await page.getByLabel('Saldo').fill('15');
    await page.locator('mat-dialog-container').getByRole('button', { name: 'Salvar' }).click();

    const linha = page.locator('tr', { hasText: codigo });
    const celulaSaldo = linha.locator('td').nth(2);
    await expect(linha).toBeVisible();
    await expect(linha.getByText('Produto criado no E2E')).toBeVisible();
    await expect(celulaSaldo).toHaveText('15');

    await linha.getByRole('button', { name: 'Editar' }).click();
    await expect(page.getByLabel('Código')).toBeDisabled();
    await page.getByLabel('Saldo').fill('40');
    await page.locator('mat-dialog-container').getByRole('button', { name: 'Salvar' }).click();
    await expect(celulaSaldo).toHaveText('40');

    page.once('dialog', (dialog) => dialog.accept());
    await linha.getByRole('button', { name: 'Excluir' }).click();
    await expect(page.locator('tr', { hasText: codigo })).toHaveCount(0);
  });

  test('sugestao via IA preenche a descricao (pula se IA indisponivel)', async ({
    page,
    request,
  }) => {
    const conta = await criarContaVerificada(request, 'e2e-ia');
    await loginViaUI(page, conta);

    await page.getByRole('button', { name: 'Novo produto' }).click();
    await page.getByLabel('Código').fill(`TEC-E2E-${Date.now()}`);

    const descricao = page.getByLabel('Descrição');
    try {
      await expect(descricao).not.toHaveValue('', { timeout: 12_000 });
    } catch {
      test.skip(true, 'IA nao respondeu a tempo (sem GEMINI_API_KEY ou indisponivel)');
    }

    await expect(descricao).not.toHaveValue('');
  });
});
