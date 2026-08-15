import { expect, test } from '@playwright/test';

import { criarContaVerificada, loginViaUI } from './helpers/auth';
import { AUTH_URL } from './helpers/urls';

test('raiz sem sessao redireciona pro login', async ({ page }) => {
  await page.goto('/');
  await page.waitForURL('**/login**');
  await expect(page.locator('mat-card-title')).toHaveText('Entrar');
});

test('cadastro cria conta e pede verificacao por e-mail', async ({ page }) => {
  const email = `e2e-cadastro-${Date.now()}@korp.local`;

  await page.goto('/cadastro');
  await page.getByLabel('E-mail').fill(email);
  await page.getByLabel('Senha').fill('senha12345');
  await page.locator('form').getByRole('button', { name: 'Cadastrar' }).click();

  await expect(page.getByText('Confira seu e-mail pra verificar a conta')).toBeVisible();
  await page.waitForURL('**/login');
});

test('login antes de verificar o e-mail mostra erro', async ({ page, request }) => {
  const email = `e2e-naoverificado-${Date.now()}@korp.local`;
  const senha = 'senha12345';
  const cadastro = await request.post(`${AUTH_URL}/auth/cadastro`, { data: { email, senha } });
  expect(cadastro.status()).toBe(201);

  await page.goto('/login');
  await page.getByLabel('E-mail').fill(email);
  await page.getByLabel('Senha').fill(senha);
  await page.locator('form').getByRole('button', { name: 'Entrar' }).click();

  await expect(page.getByText('email nao verificado')).toBeVisible();
  await expect(page).toHaveURL(/\/login/);
});

test('login apos verificar entra e mostra a navbar com o e-mail', async ({ page, request }) => {
  const conta = await criarContaVerificada(request, 'e2e-login');

  await loginViaUI(page, conta);

  await expect(page).toHaveURL(/\/produtos/);
  await expect(page.getByText(conta.email)).toBeVisible();
});

test('sair volta pro login e bloqueia rota protegida de novo', async ({ page, request }) => {
  const conta = await criarContaVerificada(request, 'e2e-logout');
  await loginViaUI(page, conta);

  await page.getByRole('button', { name: 'Sair' }).click();
  await page.waitForURL('**/login');

  await page.goto('/notas');
  await page.waitForURL('**/login**');
});

test('returnUrl leva de volta pra rota original depois do login', async ({ page, request }) => {
  const conta = await criarContaVerificada(request, 'e2e-returnurl');

  await page.goto('/notas');
  await page.waitForURL('**/login**');
  expect(page.url()).toContain('returnUrl=%2Fnotas');

  await page.getByLabel('E-mail').fill(conta.email);
  await page.getByLabel('Senha').fill(conta.senha);
  await page.locator('form').getByRole('button', { name: 'Entrar' }).click();

  await page.waitForURL('**/notas');
});
