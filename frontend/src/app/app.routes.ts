import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'produtos', pathMatch: 'full' },
  {
    path: 'produtos',
    loadComponent: () =>
      import('./features/produtos/produtos-lista/produtos-lista').then((m) => m.ProdutosLista),
  },
  {
    path: 'notas',
    loadComponent: () =>
      import('./features/notas/notas-lista/notas-lista').then((m) => m.NotasLista),
  },
  {
    path: 'notas/:id',
    loadComponent: () =>
      import('./features/notas/nota-detalhe/nota-detalhe').then((m) => m.NotaDetalhe),
  },
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login').then((m) => m.Login),
  },
  {
    path: 'cadastro',
    loadComponent: () => import('./features/auth/cadastro/cadastro').then((m) => m.Cadastro),
  },
  { path: '**', redirectTo: 'produtos' },
];
