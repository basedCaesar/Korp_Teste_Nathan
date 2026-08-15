import { execSync } from 'node:child_process';
import path from 'node:path';

import { ESTOQUE_URL } from './urls';

const REPO_ROOT = path.resolve(__dirname, '..', '..', '..');

export function pararEstoque() {
  execSync('docker compose stop estoque', { cwd: REPO_ROOT, stdio: 'ignore' });
}

export function iniciarEstoque() {
  execSync('docker compose start estoque', { cwd: REPO_ROOT, stdio: 'ignore' });
}

export async function esperarEstoqueSaudavel(timeoutMs = 30_000) {
  const inicio = Date.now();
  while (Date.now() - inicio < timeoutMs) {
    try {
      const resp = await fetch(`${ESTOQUE_URL}/health`);
      if (resp.ok) {
        return;
      }
    } catch {
      // servico ainda subindo
    }
    await new Promise((resolve) => setTimeout(resolve, 500));
  }
  throw new Error('estoque nao ficou saudavel a tempo');
}
