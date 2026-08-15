import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { Observable, tap } from 'rxjs';

import { ConfigService } from '../../core/config/config-service';
import { CadastroInput, CadastroResponse, LoginInput, LoginResponse, Usuario } from '../../core/models/auth';

const CHAVE_TOKEN = 'korp.token';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(ConfigService);

  private readonly usuarioSignal = signal<Usuario | null>(this.usuarioDoTokenSalvo());
  readonly usuario = this.usuarioSignal.asReadonly();
  readonly logado = computed(() => this.usuarioSignal() !== null);

  private get baseUrl(): string {
    return this.config.get().authUrl;
  }

  cadastrar(input: CadastroInput): Observable<CadastroResponse> {
    return this.http.post<CadastroResponse>(`${this.baseUrl}/auth/cadastro`, input);
  }

  login(input: LoginInput): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(`${this.baseUrl}/auth/login`, input).pipe(
      tap((resposta) => {
        localStorage.setItem(CHAVE_TOKEN, resposta.token);
        this.usuarioSignal.set(this.decodificarToken(resposta.token));
      }),
    );
  }

  logout(): void {
    localStorage.removeItem(CHAVE_TOKEN);
    this.usuarioSignal.set(null);
  }

  token(): string | null {
    return localStorage.getItem(CHAVE_TOKEN);
  }

  private usuarioDoTokenSalvo(): Usuario | null {
    const token = localStorage.getItem(CHAVE_TOKEN);
    if (!token) {
      return null;
    }
    const usuario = this.decodificarToken(token);
    if (!usuario) {
      localStorage.removeItem(CHAVE_TOKEN);
    }
    return usuario;
  }

  private decodificarToken(token: string): Usuario | null {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      if (typeof payload.exp === 'number' && payload.exp * 1000 < Date.now()) {
        return null;
      }
      return { user_id: payload.user_id, email: payload.email };
    } catch {
      return null;
    }
  }
}
