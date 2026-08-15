import { Component, OnInit, inject, signal } from '@angular/core';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { finalize } from 'rxjs';

import { Produto } from '../../../core/models/produto';
import { ItemNota } from '../../../core/models/nota';
import { ProdutoService } from '../../produtos/produto-service';
import { NotaService } from '../nota-service';

export interface ItemFormDialogData {
  notaId: number;
  item: ItemNota | null;
}

@Component({
  selector: 'app-item-form-dialog',
  imports: [
    ReactiveFormsModule,
    MatDialogModule,
    MatFormFieldModule,
    MatInputModule,
    MatAutocompleteModule,
    MatButtonModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './item-form-dialog.html',
  styleUrl: './item-form-dialog.scss',
})
export class ItemFormDialog implements OnInit {
  private readonly dialogRef = inject(MatDialogRef<ItemFormDialog>);
  private readonly produtoService = inject(ProdutoService);
  private readonly notaService = inject(NotaService);
  protected readonly data = inject<ItemFormDialogData>(MAT_DIALOG_DATA);

  protected readonly editando = this.data.item !== null;
  protected readonly salvando = signal(false);
  protected readonly produtos = signal<Produto[]>([]);
  protected produtoSelecionado: Produto | null = null;

  protected readonly form = new FormGroup({
    produtoBusca: new FormControl(this.data.item?.produto_codigo ?? '', {
      nonNullable: true,
      validators: [Validators.required],
    }),
    quantidade: new FormControl(this.data.item?.quantidade ?? 1, {
      nonNullable: true,
      validators: [Validators.required, Validators.min(1)],
    }),
  });

  ngOnInit(): void {
    if (!this.editando) {
      this.produtoService.listar().subscribe((produtos) => this.produtos.set(produtos));
    } else {
      this.form.controls.produtoBusca.disable();
    }
  }

  protected produtosFiltrados(): Produto[] {
    const termo = this.form.controls.produtoBusca.value.toLowerCase();
    return this.produtos().filter(
      (produto) =>
        produto.codigo.toLowerCase().includes(termo) ||
        produto.descricao.toLowerCase().includes(termo),
    );
  }

  protected selecionarProduto(produto: Produto): void {
    this.produtoSelecionado = produto;
    this.form.controls.produtoBusca.setValue(`${produto.codigo} — ${produto.descricao}`);
  }

  protected salvar(): void {
    if (this.form.invalid || (!this.editando && !this.produtoSelecionado)) {
      this.form.markAllAsTouched();
      return;
    }

    this.salvando.set(true);
    const quantidade = this.form.controls.quantidade.value;

    const requisicao = this.editando
      ? this.notaService.atualizarItem(this.data.notaId, this.data.item!.id, { quantidade })
      : this.notaService.adicionarItem(this.data.notaId, {
          produto_id: this.produtoSelecionado!.id,
          produto_codigo: this.produtoSelecionado!.codigo,
          produto_descricao: this.produtoSelecionado!.descricao,
          quantidade,
        });

    requisicao.pipe(finalize(() => this.salvando.set(false))).subscribe({
      next: (item) => this.dialogRef.close(item),
    });
  }

  protected cancelar(): void {
    this.dialogRef.close(null);
  }
}
