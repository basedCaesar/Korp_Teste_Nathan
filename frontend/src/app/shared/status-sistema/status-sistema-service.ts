import { HttpClient, HttpContext } from '@angular/common/http';
import { Injectable, inject, signal } from '@angular/core';
import { Subscription, catchError, interval, map, of, startWith, switchMap } from 'rxjs';

import { ConfigService } from '../../core/config/config-service';
import { IGNORAR_ERRO_GLOBAL } from '../../core/interceptors/error-envelope-interceptor';

interface HealthDependencias {
  dependencias: Record<string, boolean>;
}

@Injectable({ providedIn: 'root' })
export class StatusSistemaService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(ConfigService);

  readonly estoqueDisponivel = signal(true);

  private subscription: Subscription | null = null;

  iniciar(): void {
    if (this.subscription) {
      return;
    }
    this.subscription = interval(5000)
      .pipe(
        startWith(0),
        switchMap(() => this.verificarEstoque()),
      )
      .subscribe((disponivel) => this.estoqueDisponivel.set(disponivel));
  }

  parar(): void {
    this.subscription?.unsubscribe();
    this.subscription = null;
  }

  private verificarEstoque() {
    const context = new HttpContext().set(IGNORAR_ERRO_GLOBAL, true);
    return this.http
      .get<HealthDependencias>(`${this.config.get().faturamentoUrl}/health/dependencias`, { context })
      .pipe(
        map((resposta) => resposta.dependencias['estoque'] ?? false),
        catchError(() => of(false)),
      );
  }
}
