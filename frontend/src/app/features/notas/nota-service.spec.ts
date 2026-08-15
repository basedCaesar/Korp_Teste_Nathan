import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';

import { ConfigService } from '../../core/config/config-service';
import { NotaService } from './nota-service';

describe('NotaService', () => {
  let service: NotaService;
  let http: HttpTestingController;

  const configStub = { get: () => ({ faturamentoUrl: 'http://faturamento.test' }) };

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        { provide: ConfigService, useValue: configStub },
      ],
    });
    service = TestBed.inject(NotaService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('cria nota com POST vazio', () => {
    service.criar().subscribe();

    const req = http.expectOne('http://faturamento.test/notas');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({});
    req.flush({});
  });

  it('adiciona item com o payload certo', () => {
    service
      .adicionarItem(3, {
        produto_id: 1,
        produto_codigo: 'P001',
        produto_descricao: 'Parafuso',
        quantidade: 5,
      })
      .subscribe();

    const req = http.expectOne('http://faturamento.test/notas/3/itens');
    expect(req.request.method).toBe('POST');
    expect(req.request.body).toEqual({
      produto_id: 1,
      produto_codigo: 'P001',
      produto_descricao: 'Parafuso',
      quantidade: 5,
    });
    req.flush({});
  });

  it('imprimir manda o header Idempotency-Key', () => {
    service.imprimir(3, 'chave-123').subscribe();

    const req = http.expectOne('http://faturamento.test/notas/3/imprimir');
    expect(req.request.method).toBe('POST');
    expect(req.request.headers.get('Idempotency-Key')).toBe('chave-123');
    req.flush(null);
  });

  it('remove item por id', () => {
    service.removerItem(3, 9).subscribe();

    const req = http.expectOne('http://faturamento.test/notas/3/itens/9');
    expect(req.request.method).toBe('DELETE');
    req.flush(null);
  });

  it('exclui nota por id', () => {
    service.excluir(3).subscribe();

    const req = http.expectOne('http://faturamento.test/notas/3');
    expect(req.request.method).toBe('DELETE');
    req.flush(null);
  });
});
