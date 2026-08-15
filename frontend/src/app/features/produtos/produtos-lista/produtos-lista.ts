import { Component, OnInit, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { finalize } from 'rxjs';

import { Produto } from '../../../core/models/produto';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { ProdutoFormDialog, ProdutoFormDialogData } from '../produto-form-dialog/produto-form-dialog';
import { ProdutoService } from '../produto-service';

@Component({
  selector: 'app-produtos-lista',
  imports: [MatTableModule, MatButtonModule, MatIconModule, MatDialogModule, MatProgressSpinnerModule],
  templateUrl: './produtos-lista.html',
  styleUrl: './produtos-lista.scss',
})
export class ProdutosLista implements OnInit {
  private readonly produtoService = inject(ProdutoService);
  private readonly notificacao = inject(NotificacaoService);
  private readonly dialog = inject(MatDialog);

  protected readonly colunas = ['codigo', 'descricao', 'saldo', 'categoria', 'acoes'];
  protected readonly produtos = signal<Produto[]>([]);
  protected readonly carregando = signal(false);

  ngOnInit(): void {
    this.carregar();
  }

  protected carregar(): void {
    this.carregando.set(true);
    this.produtoService
      .listar()
      .pipe(finalize(() => this.carregando.set(false)))
      .subscribe((produtos) => this.produtos.set(produtos));
  }

  protected novo(): void {
    this.abrirFormulario(null);
  }

  protected editar(produto: Produto): void {
    this.abrirFormulario(produto);
  }

  private abrirFormulario(produto: Produto | null): void {
    const ref = this.dialog.open<ProdutoFormDialog, ProdutoFormDialogData, Produto | null>(
      ProdutoFormDialog,
      { data: { produto } },
    );
    ref.afterClosed().subscribe((resultado) => {
      if (resultado) {
        this.carregar();
      }
    });
  }

  protected excluir(produto: Produto): void {
    if (!confirm(`Excluir o produto ${produto.codigo}?`)) {
      return;
    }
    this.produtoService.excluir(produto.id).subscribe(() => {
      this.notificacao.sucesso('Produto excluído.');
      this.carregar();
    });
  }
}
