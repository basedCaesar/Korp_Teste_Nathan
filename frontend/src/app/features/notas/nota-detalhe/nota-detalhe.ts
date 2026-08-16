import { Component, OnDestroy, OnInit, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { ActivatedRoute, Router } from '@angular/router';
import { finalize } from 'rxjs';

import { ItemNota, Nota } from '../../../core/models/nota';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { StatusSistemaService } from '../../../shared/status-sistema/status-sistema-service';
import { ItemFormDialog, ItemFormDialogData } from '../item-form-dialog/item-form-dialog';
import { NotaService } from '../nota-service';

@Component({
  selector: 'app-nota-detalhe',
  imports: [MatTableModule, MatButtonModule, MatIconModule, MatProgressSpinnerModule],
  templateUrl: './nota-detalhe.html',
  styleUrl: './nota-detalhe.scss',
})
export class NotaDetalhe implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly notaService = inject(NotaService);
  private readonly notificacao = inject(NotificacaoService);
  private readonly dialog = inject(MatDialog);
  private readonly statusSistema = inject(StatusSistemaService);

  protected readonly colunasItens = ['produto', 'quantidade', 'acoes'];
  protected readonly nota = signal<Nota | null>(null);
  protected readonly carregando = signal(false);
  protected readonly imprimindo = signal(false);
  protected readonly estoqueDisponivel = this.statusSistema.estoqueDisponivel;

  private get notaId(): number {
    return Number(this.route.snapshot.paramMap.get('id'));
  }

  ngOnInit(): void {
    this.carregar();
    this.statusSistema.iniciar();
  }

  ngOnDestroy(): void {
    this.statusSistema.parar();
  }

  protected carregar(): void {
    this.carregando.set(true);
    this.notaService
      .buscar(this.notaId)
      .pipe(finalize(() => this.carregando.set(false)))
      .subscribe((nota) => this.nota.set(nota));
  }

  protected adicionarItem(): void {
    this.abrirItemForm(null);
  }

  protected editarItem(item: ItemNota): void {
    this.abrirItemForm(item);
  }

  private abrirItemForm(item: ItemNota | null): void {
    const ref = this.dialog.open<ItemFormDialog, ItemFormDialogData, ItemNota | null>(ItemFormDialog, {
      data: { notaId: this.notaId, item },
    });
    ref.afterClosed().subscribe((resultado) => {
      if (resultado) {
        this.carregar();
      }
    });
  }

  protected removerItem(item: ItemNota): void {
    if (!confirm(`Remover ${item.produto_codigo} da nota?`)) {
      return;
    }
    this.notaService.removerItem(this.notaId, item.id).subscribe(() => {
      this.notificacao.sucesso('Item removido.');
      this.carregar();
    });
  }

  protected imprimir(): void {
    const idempotencyKey = crypto.randomUUID();
    this.imprimindo.set(true);
    this.notaService
      .imprimir(this.notaId, idempotencyKey)
      .pipe(finalize(() => this.imprimindo.set(false)))
      .subscribe({
        next: () => {
          this.notificacao.sucesso('Nota fiscal impressa com sucesso.');
          this.carregar();
        },
        error: () => this.carregar(),
      });
  }

  protected excluirNota(): void {
    if (!confirm('Excluir esta nota?')) {
      return;
    }
    this.notaService.excluir(this.notaId).subscribe(() => {
      this.notificacao.sucesso('Nota excluída.');
      this.router.navigate(['/notas']);
    });
  }
}
