import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable, catchError, debounceTime, distinctUntilChanged, of, switchMap } from 'rxjs';

import { ConfigService } from '../../core/config/config-service';
import { Produto, ProdutoInput, ProdutoUpdateInput, SugestaoProduto } from '../../core/models/produto';

@Injectable({ providedIn: 'root' })
export class ProdutoService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(ConfigService);

  private get baseUrl(): string {
    return `${this.config.get().estoqueUrl}/produtos`;
  }

  listar(): Observable<Produto[]> {
    return this.http.get<Produto[]>(this.baseUrl);
  }

  buscar(id: number): Observable<Produto> {
    return this.http.get<Produto>(`${this.baseUrl}/${id}`);
  }

  criar(input: ProdutoInput): Observable<Produto> {
    return this.http.post<Produto>(this.baseUrl, input);
  }

  atualizar(id: number, input: ProdutoUpdateInput): Observable<Produto> {
    return this.http.put<Produto>(`${this.baseUrl}/${id}`, input);
  }

  excluir(id: number): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  sugerir(codigo: string, categoria: string): Observable<SugestaoProduto> {
    return this.http.post<SugestaoProduto>(`${this.baseUrl}/sugestao`, { codigo, categoria });
  }

  sugestaoAoDigitar(
    entrada$: Observable<{ codigo: string; categoria: string }>,
  ): Observable<SugestaoProduto | null> {
    return entrada$.pipe(
      debounceTime(600),
      distinctUntilChanged((a, b) => a.codigo === b.codigo && a.categoria === b.categoria),
      switchMap(({ codigo, categoria }) =>
        this.sugerir(codigo, categoria).pipe(catchError(() => of(null))),
      ),
    );
  }
}
