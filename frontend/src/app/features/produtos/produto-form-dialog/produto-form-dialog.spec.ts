import { TestBed } from '@angular/core/testing';
import { MAT_DIALOG_DATA, MatDialogRef } from '@angular/material/dialog';
import { of } from 'rxjs';

import { Produto } from '../../../core/models/produto';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { ProdutoService } from '../produto-service';
import { ProdutoFormDialog, ProdutoFormDialogData } from './produto-form-dialog';

describe('ProdutoFormDialog', () => {
  let produtoServiceMock: {
    criar: ReturnType<typeof vi.fn>;
    atualizar: ReturnType<typeof vi.fn>;
    sugestaoAoDigitar: ReturnType<typeof vi.fn>;
    listar: ReturnType<typeof vi.fn>;
  };
  let dialogRefMock: { close: ReturnType<typeof vi.fn> };
  let notificacaoMock: { sucesso: ReturnType<typeof vi.fn>; erro: ReturnType<typeof vi.fn> };

  function montar(data: ProdutoFormDialogData) {
    TestBed.configureTestingModule({
      imports: [ProdutoFormDialog],
      providers: [
        { provide: ProdutoService, useValue: produtoServiceMock },
        { provide: NotificacaoService, useValue: notificacaoMock },
        { provide: MatDialogRef, useValue: dialogRefMock },
        { provide: MAT_DIALOG_DATA, useValue: data },
      ],
    });
    const fixture = TestBed.createComponent(ProdutoFormDialog);
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => {
    produtoServiceMock = {
      criar: vi.fn(),
      atualizar: vi.fn(),
      sugestaoAoDigitar: vi.fn().mockReturnValue(of()),
      listar: vi.fn().mockReturnValue(of([])),
    };
    dialogRefMock = { close: vi.fn() };
    notificacaoMock = { sucesso: vi.fn(), erro: vi.fn() };
  });

  it('nao chama o service se o form estiver invalido', () => {
    const fixture = montar({ produto: null });
    fixture.componentInstance['form'].patchValue({ codigo: '', descricao: '', saldo: 0 });

    fixture.componentInstance['salvar']();

    expect(produtoServiceMock.criar).not.toHaveBeenCalled();
    expect(dialogRefMock.close).not.toHaveBeenCalled();
  });

  it('cria produto novo com o payload certo e fecha o dialog', () => {
    const produtoCriado: Produto = {
      id: 1,
      codigo: 'P010',
      descricao: 'Chave de fenda',
      saldo: 20,
      categoria: 'Ferramentas',
      version: 0,
      created_at: '',
      updated_at: '',
    };
    produtoServiceMock.criar.mockReturnValue(of(produtoCriado));

    const fixture = montar({ produto: null });
    fixture.componentInstance['form'].setValue({
      codigo: 'P010',
      descricao: 'Chave de fenda',
      saldo: 20,
      categoria: 'Ferramentas',
    });

    fixture.componentInstance['salvar']();

    expect(produtoServiceMock.criar).toHaveBeenCalledWith({
      codigo: 'P010',
      descricao: 'Chave de fenda',
      saldo: 20,
      categoria: 'Ferramentas',
    });
    expect(dialogRefMock.close).toHaveBeenCalledWith(produtoCriado);
  });

  it('edita produto existente sem reenviar o codigo', () => {
    const produtoExistente: Produto = {
      id: 5,
      codigo: 'P020',
      descricao: 'Martelo',
      saldo: 4,
      categoria: 'Ferramentas',
      version: 2,
      created_at: '',
      updated_at: '',
    };
    produtoServiceMock.atualizar.mockReturnValue(of({ ...produtoExistente, saldo: 9 }));

    const fixture = montar({ produto: produtoExistente });
    fixture.componentInstance['form'].patchValue({ saldo: 9 });

    fixture.componentInstance['salvar']();

    expect(produtoServiceMock.atualizar).toHaveBeenCalledWith(5, {
      descricao: 'Martelo',
      saldo: 9,
      categoria: 'Ferramentas',
    });
  });

  it('cancelar fecha o dialog sem resultado', () => {
    const fixture = montar({ produto: null });

    fixture.componentInstance['cancelar']();

    expect(dialogRefMock.close).toHaveBeenCalledWith(null);
  });

  it('carrega as categorias distintas ja usadas pelo usuario', () => {
    produtoServiceMock.listar.mockReturnValue(
      of([
        { id: 1, codigo: 'A', descricao: '', saldo: 0, categoria: 'Perifericos', version: 0, created_at: '', updated_at: '' },
        { id: 2, codigo: 'B', descricao: '', saldo: 0, categoria: 'Perifericos', version: 0, created_at: '', updated_at: '' },
        { id: 3, codigo: 'C', descricao: '', saldo: 0, categoria: 'Cabos', version: 0, created_at: '', updated_at: '' },
        { id: 4, codigo: 'D', descricao: '', saldo: 0, categoria: '', version: 0, created_at: '', updated_at: '' },
      ]),
    );

    const fixture = montar({ produto: null });

    expect(fixture.componentInstance['categoriasExistentes']()).toEqual(['Perifericos', 'Cabos']);
  });
});
