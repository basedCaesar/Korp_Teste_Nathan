import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { Subject } from 'rxjs';

import { ConfigService } from '../../core/config/config-service';
import { Produto } from '../../core/models/produto';
import { ProdutoService } from './produto-service';

describe('ProdutoService', () => {
  let service: ProdutoService;
  let http: HttpTestingController;

  const configStub = { get: () => ({ estoqueUrl: 'http://estoque.test' }) };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ConfigService, useValue: configStub },
      ],
    });
    service = TestBed.inject(ProdutoService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('lista produtos do endpoint do estoque configurado', () => {
    const produtos: Produto[] = [
      {
        id: 1,
        codigo: 'P001',
        descricao: 'Parafuso',
        saldo: 10,
        version: 0,
        created_at: '',
        updated_at: '',
      },
    ];

    service.listar().subscribe((resultado) => {
      expect(resultado).toEqual(produtos);
    });

    const req = http.expectOne('http://estoque.test/produtos');
    expect(req.request.method).toBe('GET');
    req.flush(produtos);
  });

  it('cria produto com POST no corpo certo', () => {
    service.criar({ codigo: 'P002', descricao: 'Porca', saldo: 5 }).subscribe();

    const req = http.expectOne('http://estoque.test/produtos');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ codigo: 'P002', descricao: 'Porca', saldo: 5 });
    req.flush({});
  });

  it('atualiza produto sem enviar o codigo (imutavel)', () => {
    service.atualizar(7, { descricao: 'Nova descricao', saldo: 20 }).subscribe();

    const req = http.expectOne('http://estoque.test/produtos/7');
    expect(req.request.method).toBe('PUT');
    expect(req.request.body).toEqual({ descricao: 'Nova descricao', saldo: 20 });
    req.flush({});
  });

  it('exclui produto por id', () => {
    service.excluir(9).subscribe();

    const req = http.expectOne('http://estoque.test/produtos/9');
    expect(req.request.method).toBe('DELETE');
    req.flush(null);
  });

  it('pede sugestao via IA com o codigo informado', () => {
    service.sugerir('TEC-045').subscribe();

    const req = http.expectOne('http://estoque.test/produtos/sugestao');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({ codigo: 'TEC-045' });
    req.flush({ codigo: 'TEC-045', descricao_sugerida: '', produtos_similares: [] });
  });

  it('sugestaoAoDigitar so chama a API depois do debounce, e so 1x para digitacao repetida', () => {
    vi.useFakeTimers();
    const codigo$ = new Subject<string>();
    const emissoes: unknown[] = [];

    service.sugestaoAoDigitar(codigo$).subscribe((valor) => emissoes.push(valor));

    codigo$.next('T');
    codigo$.next('TE');
    codigo$.next('TEC');
    vi.advanceTimersByTime(599);

    http.expectNone((req) => req.url.endsWith('/produtos/sugestao'));
    vi.advanceTimersByTime(1);

    const req = http.expectOne('http://estoque.test/produtos/sugestao');
    expect(req.request.body).toEqual({ codigo: 'TEC' });
    req.flush({ codigo: 'TEC', descricao_sugerida: 'x', produtos_similares: [] });

    expect(emissoes.length).toBe(1);
    vi.useRealTimers();
  });

  it('sugestaoAoDigitar sobrevive a uma falha e ainda responde ao proximo codigo digitado', () => {
    vi.useFakeTimers();
    const codigo$ = new Subject<string>();
    const emissoes: unknown[] = [];

    service.sugestaoAoDigitar(codigo$).subscribe((valor) => emissoes.push(valor));

    codigo$.next('FALHA');
    vi.advanceTimersByTime(600);
    http
      .expectOne('http://estoque.test/produtos/sugestao')
      .flush(
        { code: 'IA_INDISPONIVEL', message: 'servico de IA indisponivel', details: [], trace_id: '' },
        { status: 503, statusText: 'Service Unavailable' },
      );

    expect(emissoes).toEqual([null]);

    codigo$.next('OK');
    vi.advanceTimersByTime(600);
    const segundaReq = http.expectOne('http://estoque.test/produtos/sugestao');
    segundaReq.flush({ codigo: 'OK', descricao_sugerida: 'Produto OK', produtos_similares: [] });

    expect(emissoes.length).toBe(2);
    expect((emissoes[1] as { descricao_sugerida: string }).descricao_sugerida).toBe('Produto OK');
    vi.useRealTimers();
  });
});
