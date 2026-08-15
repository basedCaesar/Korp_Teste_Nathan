import { APIRequestContext, Page, expect } from '@playwright/test';

import { AUTH_URL, MAILHOG_URL } from './urls';

export interface ContaTeste {
  email: string;
  senha: string;
  token: string;
}

export async function criarContaVerificada(
  request: APIRequestContext,
  prefixo: string,
): Promise<ContaTeste> {
  const email = `${prefixo}-${Date.now()}-${Math.floor(Math.random() * 100000)}@korp.local`;
  const senha = 'senha12345';

  const cadastro = await request.post(`${AUTH_URL}/auth/cadastro`, { data: { email, senha } });
  expect(cadastro.status(), await cadastro.text()).toBe(201);

  const tokenVerificacao = await esperarTokenVerificacao(request, email);
  const verificar = await request.get(`${AUTH_URL}/auth/verificar?token=${tokenVerificacao}`);
  expect(verificar.status()).toBe(200);

  const login = await request.post(`${AUTH_URL}/auth/login`, { data: { email, senha } });
  expect(login.status(), await login.text()).toBe(200);
  const corpo = await login.json();

  return { email, senha, token: corpo.token };
}

async function esperarTokenVerificacao(
  request: APIRequestContext,
  email: string,
): Promise<string> {
  for (let tentativa = 0; tentativa < 15; tentativa++) {
    const resp = await request.get(`${MAILHOG_URL}/api/v2/search?kind=to&query=${email}`);
    if (resp.ok()) {
      const corpo = await resp.json();
      const item = corpo.items?.[0];
      if (item) {
        const match = (item.Content.Body as string).match(/token=([a-f0-9]+)/);
        if (match) {
          return match[1];
        }
      }
    }
    await new Promise((resolve) => setTimeout(resolve, 1000));
  }
  throw new Error(`token de verificacao nao encontrado no MailHog pra ${email}`);
}

export async function loginViaUI(page: Page, conta: Pick<ContaTeste, 'email' | 'senha'>) {
  await page.goto('/login');
  await page.getByLabel('E-mail').fill(conta.email);
  await page.getByLabel('Senha').fill(conta.senha);
  await page.locator('form').getByRole('button', { name: 'Entrar' }).click();
  await page.waitForURL('**/produtos');
}
