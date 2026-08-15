import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap } from '@angular/router';
import { Subject, of, throwError } from 'rxjs';

import { Nota } from '../../../core/models/nota';
import { NotificacaoService } from '../../../shared/notificacao/notificacao-service';
import { NotaService } from '../nota-service';
import { NotaDetalhe } from './nota-detalhe';

// Assim como em ProdutosLista, abrir o MatDialog (adicionar/editar item) fica pro E2E — aqui
// o foco é o que é próprio do componente: carga da nota e, principalmente, o fluxo de
// impressão (chave de idempotência por clique, loading, recarga em sucesso e em erro).
describe('NotaDetalhe', () => {
  const notaAberta: Nota = {
    id: 3,
    numero: 3,
    status: 'ABERTA',
    itens: [],
    created_at: '',
    updated_at: '',
  };

  let notaServiceMock: {
    buscar: ReturnType<typeof vi.fn>;
    imprimir: ReturnType<typeof vi.fn>;
    excluir: ReturnType<typeof vi.fn>;
  };
  let notificacaoMock: { sucesso: ReturnType<typeof vi.fn>; erro: ReturnType<typeof vi.fn> };
  let router: { navigate: ReturnType<typeof vi.fn> };

  function montar() {
    TestBed.configureTestingModule({
      imports: [NotaDetalhe],
      providers: [
        { provide: NotaService, useValue: notaServiceMock },
        { provide: NotificacaoService, useValue: notificacaoMock },
        { provide: Router, useValue: router },
        {
          provide: ActivatedRoute,
          useValue: { snapshot: { paramMap: convertToParamMap({ id: '3' }) } },
        },
      ],
    });
    const fixture = TestBed.createComponent(NotaDetalhe);
    fixture.detectChanges();
    return fixture;
  }

  beforeEach(() => {
    notaServiceMock = {
      buscar: vi.fn().mockReturnValue(of(notaAberta)),
      imprimir: vi.fn().mockReturnValue(of(undefined)),
      excluir: vi.fn().mockReturnValue(of(undefined)),
    };
    notificacaoMock = { sucesso: vi.fn(), erro: vi.fn() };
    router = { navigate: vi.fn() };
  });

  it('carrega a nota pelo id da rota', () => {
    const fixture = montar();

    expect(notaServiceMock.buscar).toHaveBeenCalledWith(3);
    expect(fixture.componentInstance['nota']()).toEqual(notaAberta);
  });

  it('imprimir gera uma Idempotency-Key nova a cada clique e recarrega a nota no sucesso', () => {
    const fixture = montar();
    notaServiceMock.buscar.mockClear();

    fixture.componentInstance['imprimir']();
    fixture.componentInstance['imprimir']();

    expect(notaServiceMock.imprimir).toHaveBeenCalledTimes(2);
    const [id1, chave1] = notaServiceMock.imprimir.mock.calls[0];
    const [id2, chave2] = notaServiceMock.imprimir.mock.calls[1];
    expect(id1).toBe(3);
    expect(id2).toBe(3);
    expect(chave1).not.toBe(chave2);
    expect(notaServiceMock.buscar).toHaveBeenCalledTimes(2);
  });

  it('desliga o loading e recarrega a nota mesmo quando imprimir falha', () => {
    notaServiceMock.imprimir.mockReturnValue(throwError(() => new Error('estoque indisponivel')));
    const fixture = montar();
    notaServiceMock.buscar.mockClear();

    fixture.componentInstance['imprimir']();

    expect(fixture.componentInstance['imprimindo']()).toBe(false);
    expect(notaServiceMock.buscar).toHaveBeenCalledWith(3);
  });

  it('liga o loading enquanto a chamada de impressao esta pendente', () => {
    const chamada$ = new Subject<void>();
    notaServiceMock.imprimir.mockReturnValue(chamada$);
    const fixture = montar();

    fixture.componentInstance['imprimir']();
    expect(fixture.componentInstance['imprimindo']()).toBe(true);

    chamada$.next();
    chamada$.complete();
    expect(fixture.componentInstance['imprimindo']()).toBe(false);
  });

  it('exclui a nota e navega de volta pra lista quando confirmado', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    const fixture = montar();

    fixture.componentInstance['excluirNota']();

    expect(notaServiceMock.excluir).toHaveBeenCalledWith(3);
    expect(router.navigate).toHaveBeenCalledWith(['/notas']);
  });

  it('nao exclui se o usuario cancelar a confirmacao', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    const fixture = montar();

    fixture.componentInstance['excluirNota']();

    expect(notaServiceMock.excluir).not.toHaveBeenCalled();
  });
});
