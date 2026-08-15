import { DatePipe } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatTableModule } from '@angular/material/table';
import { Router } from '@angular/router';
import { finalize } from 'rxjs';

import { Nota } from '../../../core/models/nota';
import { NotaService } from '../nota-service';

@Component({
  selector: 'app-notas-lista',
  imports: [
    DatePipe,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
  ],
  templateUrl: './notas-lista.html',
  styleUrl: './notas-lista.scss',
})
export class NotasLista implements OnInit {
  private readonly notaService = inject(NotaService);
  private readonly router = inject(Router);

  protected readonly colunas = ['numero', 'status', 'criada_em'];
  protected readonly notas = signal<Nota[]>([]);
  protected readonly carregando = signal(false);
  protected readonly criando = signal(false);

  ngOnInit(): void {
    this.carregar();
  }

  protected carregar(): void {
    this.carregando.set(true);
    this.notaService
      .listar()
      .pipe(finalize(() => this.carregando.set(false)))
      .subscribe((notas) => this.notas.set(notas));
  }

  protected nova(): void {
    this.criando.set(true);
    this.notaService
      .criar()
      .pipe(finalize(() => this.criando.set(false)))
      .subscribe((nota) => this.router.navigate(['/notas', nota.id]));
  }

  protected abrir(nota: Nota): void {
    this.router.navigate(['/notas', nota.id]);
  }
}
