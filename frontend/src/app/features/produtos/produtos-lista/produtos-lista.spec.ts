import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';

import { Produto } from '../../../core/models/produto';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { ProdutoService } from '../produto-service';
import { ProdutosLista } from './produtos-lista';

// A abertura do MatDialog (novo/editar) é comportamento de integração com overlay real do
// Angular Material — coberto pelo E2E (Playwright, bloco F6), não faz sentido mockar os
// internals do CDK aqui. Este spec cobre a lógica própria do componente: carga da lista e
// o fluxo de exclusão (confirmação + chamada ao service).
describe('ProdutosLista', () => {
  const produtos: Produto[] = [
    { id: 1, codigo: 'P001', descricao: 'Parafuso', saldo: 10, version: 0, created_at: '', updated_at: '' },
    { id: 2, codigo: 'P002', descricao: 'Porca', saldo: 3, version: 0, created_at: '', updated_at: '' },
  ];

  let produtoServiceMock: {
    listar: ReturnType<typeof vi.fn>;
    excluir: ReturnType<typeof vi.fn>;
  };
  let notificacaoMock: { sucesso: ReturnType<typeof vi.fn>; erro: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    produtoServiceMock = {
      listar: vi.fn().mockReturnValue(of(produtos)),
      excluir: vi.fn().mockReturnValue(of(undefined)),
    };
    notificacaoMock = { sucesso: vi.fn(), erro: vi.fn() };

    TestBed.configureTestingModule({
      imports: [ProdutosLista],
      providers: [
        { provide: ProdutoService, useValue: produtoServiceMock },
        { provide: NotificacaoService, useValue: notificacaoMock },
      ],
    });
  });

  it('carrega a lista de produtos no ngOnInit', () => {
    const fixture = TestBed.createComponent(ProdutosLista);
    fixture.detectChanges();

    expect(produtoServiceMock.listar).toHaveBeenCalled();
    expect(fixture.componentInstance['produtos']()).toEqual(produtos);
  });

  it('nao exclui se o usuario cancelar a confirmacao', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    const fixture = TestBed.createComponent(ProdutosLista);
    fixture.detectChanges();

    fixture.componentInstance['excluir'](produtos[0]);

    expect(produtoServiceMock.excluir).not.toHaveBeenCalled();
  });

  it('exclui e notifica sucesso quando confirmado', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    const fixture = TestBed.createComponent(ProdutosLista);
    fixture.detectChanges();

    fixture.componentInstance['excluir'](produtos[0]);

    expect(produtoServiceMock.excluir).toHaveBeenCalledWith(produtos[0].id);
    expect(notificacaoMock.sucesso).toHaveBeenCalled();
  });

  it('recarrega a lista apos excluir', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    const fixture = TestBed.createComponent(ProdutosLista);
    fixture.detectChanges();
    produtoServiceMock.listar.mockClear();

    fixture.componentInstance['excluir'](produtos[1]);

    expect(produtoServiceMock.listar).toHaveBeenCalled();
  });
});
