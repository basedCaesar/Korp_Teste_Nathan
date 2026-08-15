import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';

import { ConfigService } from '../../core/config/config-service';
import { AuthService } from './auth-service';

function criarToken(payload: Record<string, unknown>): string {
  const cabecalho = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const corpo = btoa(JSON.stringify(payload));
  return `${cabecalho}.${corpo}.assinatura-fake`;
}

describe('AuthService', () => {
  let http: HttpTestingController;

  const configStub = { get: () => ({ authUrl: 'http://auth.test' }) };

  function configurar() {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ConfigService, useValue: configStub },
      ],
    });
    http = TestBed.inject(HttpTestingController);
    return TestBed.inject(AuthService);
  }

  afterEach(() => {
    localStorage.clear();
  });

  it('comeca deslogado quando nao ha token salvo', () => {
    const service = configurar();
    expect(service.logado()).toBe(false);
    expect(service.usuario()).toBeNull();
  });

  it('login salva o token e popula o usuario a partir do payload do JWT', () => {
    const service = configurar();
    const tokenValido = criarToken({
      user_id: 7,
      email: 'teste@korp.local',
      exp: Math.floor(Date.now() / 1000) + 3600,
    });

    service.login({ email: 'teste@korp.local', senha: 'senha123' }).subscribe();
    const req = http.expectOne('http://auth.test/auth/login');
    req.flush({ token: tokenValido });

    expect(service.logado()).toBe(true);
    expect(service.usuario()).toEqual({ user_id: 7, email: 'teste@korp.local' });
    expect(localStorage.getItem('korp.token')).toBe(tokenValido);
  });

  it('logout limpa o token e o usuario', () => {
    const service = configurar();
    const tokenValido = criarToken({
      user_id: 1,
      email: 'x@korp.local',
      exp: Math.floor(Date.now() / 1000) + 3600,
    });
    service.login({ email: 'x@korp.local', senha: 'senha' }).subscribe();
    http.expectOne('http://auth.test/auth/login').flush({ token: tokenValido });

    service.logout();

    expect(service.logado()).toBe(false);
    expect(localStorage.getItem('korp.token')).toBeNull();
  });

  it('restaura o usuario do localStorage se o token salvo ainda for valido', () => {
    const tokenValido = criarToken({
      user_id: 4,
      email: 'salvo@korp.local',
      exp: Math.floor(Date.now() / 1000) + 3600,
    });
    localStorage.setItem('korp.token', tokenValido);

    const service = configurar();

    expect(service.logado()).toBe(true);
    expect(service.usuario()?.email).toBe('salvo@korp.local');
  });

  it('descarta token expirado salvo no localStorage', () => {
    const tokenExpirado = criarToken({
      user_id: 4,
      email: 'expirado@korp.local',
      exp: Math.floor(Date.now() / 1000) - 3600,
    });
    localStorage.setItem('korp.token', tokenExpirado);

    const service = configurar();

    expect(service.logado()).toBe(false);
    expect(localStorage.getItem('korp.token')).toBeNull();
  });
});
