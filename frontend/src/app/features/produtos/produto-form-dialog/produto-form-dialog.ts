import { Component, OnInit, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
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
    MatAutocompleteModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './produto-form-dialog.html',
  styleUrl: './produto-form-dialog.scss',
})
export class ProdutoFormDialog implements OnInit {
  private readonly dialogRef = inject(MatDialogRef<ProdutoFormDialog>);
  private readonly produtoService = inject(ProdutoService);
  private readonly notificacao = inject(NotificacaoService);
  protected readonly data = inject<ProdutoFormDialogData>(MAT_DIALOG_DATA);

  protected readonly editando = this.data.produto !== null;
  protected readonly salvando = signal(false);
  protected readonly sugerindo = signal(false);
  protected readonly similares = signal<SugestaoProduto['produtos_similares']>([]);
  protected readonly categoriasExistentes = signal<string[]>([]);

  private readonly entradaDigitada$ = new Subject<{ codigo: string; categoria: string }>();

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
    categoria: new FormControl(this.data.produto?.categoria ?? '', { nonNullable: true }),
  });

  ngOnInit(): void {
    if (!this.editando) {
      this.produtoService.listar().subscribe((produtos) => {
        const categorias = new Set(produtos.map((p) => p.categoria).filter((c) => c));
        this.categoriasExistentes.set([...categorias]);
      });
    }
  }

  constructor() {
    if (!this.editando) {
      this.produtoService
        .sugestaoAoDigitar(this.entradaDigitada$)
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

  protected categoriasFiltradas(): string[] {
    const termo = this.form.controls.categoria.value.toLowerCase();
    return this.categoriasExistentes().filter((categoria) =>
      categoria.toLowerCase().includes(termo),
    );
  }

  protected aoMudarCodigoOuCategoria(): void {
    if (this.editando) {
      return;
    }
    const codigo = this.form.controls.codigo.value.trim();
    if (!codigo) {
      return;
    }
    this.sugerindo.set(true);
    this.entradaDigitada$.next({
      codigo,
      categoria: this.form.controls.categoria.value.trim(),
    });
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
          categoria: valor.categoria,
        })
      : this.produtoService.criar({
          codigo: valor.codigo,
          descricao: valor.descricao,
          saldo: valor.saldo,
          categoria: valor.categoria,
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
