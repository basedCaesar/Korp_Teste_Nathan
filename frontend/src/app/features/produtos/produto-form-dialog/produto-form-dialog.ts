import { Component, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatChipsModule } from '@angular/material/chips';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { Subject, finalize } from 'rxjs';

import { Produto, SugestaoProduto } from '../../../core/models/produto';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { ProdutoService } from '../produto-service';

export interface ProdutoFormDialogData {
  produto: Produto | null;
}

@Component({
  selector: 'app-produto-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule,
    MatIconModule,
    MatChipsModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './produto-form-dialog.html',
  styleUrl: './produto-form-dialog.scss',
})
export class ProdutoFormDialog {
  private readonly dialogRef = inject(MatDialogRef<ProdutoFormDialog>);
  private readonly produtoService = inject(ProdutoService);
  private readonly notificacao = inject(NotificacaoService);
  protected readonly data = inject<ProdutoFormDialogData>(MAT_DIALOG_DATA);

  protected readonly editando = this.data.produto !== null;
  protected readonly salvando = signal(false);
  protected readonly sugerindo = signal(false);
  protected readonly similares = signal<SugestaoProduto['produtos_similares']>([]);

  private readonly codigoDigitado$ = new Subject<string>();

  protected readonly form = new FormGroup({
    codigo: new FormControl(
      { value: this.data.produto?.codigo ?? '', disabled: this.editando },
      { nonNullable: true, validators: [Validators.required] },
    ),
    descricao: new FormControl(this.data.produto?.descricao ?? '', {
      nonNullable: true,
      validators: [Validators.required],
    }),
    saldo: new FormControl(this.data.produto?.saldo ?? 0, {
      nonNullable: true,
      validators: [Validators.required, Validators.min(0)],
    }),
  });

  constructor() {
    if (!this.editando) {
      this.produtoService
        .sugestaoAoDigitar(this.codigoDigitado$)
        .pipe(takeUntilDestroyed())
        .subscribe((sugestao) => {
          this.sugerindo.set(false);
          if (!sugestao) {
            return;
          }
          if (!this.form.controls.descricao.value) {
            this.form.controls.descricao.setValue(sugestao.descricao_sugerida);
          }
          this.similares.set(sugestao.produtos_similares);
        });
    }
  }

  protected aoDigitarCodigo(codigo: string): void {
    if (this.editando || !codigo.trim()) {
      return;
    }
    this.sugerindo.set(true);
    this.codigoDigitado$.next(codigo.trim());
  }

  protected salvar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.salvando.set(true);
    const valor = this.form.getRawValue();

    const requisicao = this.editando
      ? this.produtoService.atualizar(this.data.produto!.id, {
          descricao: valor.descricao,
          saldo: valor.saldo,
        })
      : this.produtoService.criar({
          codigo: valor.codigo,
          descricao: valor.descricao,
          saldo: valor.saldo,
        });

    requisicao.pipe(finalize(() => this.salvando.set(false))).subscribe({
      next: (produto) => {
        this.notificacao.sucesso(
          this.editando ? 'Produto atualizado.' : `Produto ${produto.codigo} criado.`,
        );
        this.dialogRef.close(produto);
      },
    });
  }

  protected cancelar(): void {
    this.dialogRef.close(null);
  }
}
