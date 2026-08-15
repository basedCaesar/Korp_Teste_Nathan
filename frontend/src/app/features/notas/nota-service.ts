import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { ConfigService } from '../../core/config/config-service';
import { ItemInput, ItemNota, ItemUpdateInput, Nota } from '../../core/models/nota';

@Injectable({ providedIn: 'root' })
export class NotaService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(ConfigService);

  private get baseUrl(): string {
    return `${this.config.get().faturamentoUrl}/notas`;
  }

  listar(): Observable<Nota[]> {
    return this.http.get<Nota[]>(this.baseUrl);
  }

  buscar(id: number): Observable<Nota> {
    return this.http.get<Nota>(`${this.baseUrl}/${id}`);
  }

  criar(): Observable<Nota> {
    return this.http.post<Nota>(this.baseUrl, {});
  }

  excluir(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  adicionarItem(notaId: number, input: ItemInput): Observable<ItemNota> {
    return this.http.post<ItemNota>(`${this.baseUrl}/${notaId}/itens`, input);
  }

  atualizarItem(notaId: number, itemId: number, input: ItemUpdateInput): Observable<ItemNota> {
    return this.http.put<ItemNota>(`${this.baseUrl}/${notaId}/itens/${itemId}`, input);
  }

  removerItem(notaId: number, itemId: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${notaId}/itens/${itemId}`);
  }

  imprimir(notaId: number, idempotencyKey: string): Observable<void> {
    const headers = new HttpHeaders({ 'Idempotency-Key': idempotencyKey });
    return this.http.post<void>(`${this.baseUrl}/${notaId}/imprimir`, null, { headers });
  }
}
