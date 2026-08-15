import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { NotificacaoService } from '../../shared/notificacao/notificacao-service';
import { errorEnvelopeInterceptor } from './error-envelope-interceptor';

describe('errorEnvelopeInterceptor', () => {
  let http: HttpClient;
  let httpTesting: HttpTestingController;
  let notificacaoMock: { erro: ReturnType<typeof vi.fn>; sucesso: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    notificacaoMock = { erro: vi.fn(), sucesso: vi.fn() };

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([errorEnvelopeInterceptor])),
        provideHttpClientTesting(),
        { provide: NotificacaoService, useValue: notificacaoMock },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpTesting = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpTesting.verify();
  });

  it('mostra a message do envelope de erro do backend', () => {
    http.get('/produtos/999').subscribe({ error: () => {} });

    const req = httpTesting.expectOne('/produtos/999');
    req.flush(
      { code: 'PRODUTO_NAO_ENCONTRADO', message: 'produto nao encontrado', details: [], trace_id: 'abc' },
      { status: 404, statusText: 'Not Found' },
    );

    expect(notificacaoMock.erro).toHaveBeenCalledWith('produto nao encontrado');
  });

  it('mostra mensagem generica quando nao consegue conectar (status 0)', () => {
    http.get('/produtos').subscribe({ error: () => {} });

    const req = httpTesting.expectOne('/produtos');
    req.error(new ProgressEvent('error'), { status: 0, statusText: 'Unknown Error' });

    expect(notificacaoMock.erro).toHaveBeenCalledWith('Não foi possível conectar ao servidor.');
  });

  it('mostra mensagem generica quando o corpo do erro nao segue o envelope', () => {
    http.get('/produtos').subscribe({ error: () => {} });

    const req = httpTesting.expectOne('/produtos');
    req.flush('erro cru sem envelope', { status: 500, statusText: 'Internal Server Error' });

    expect(notificacaoMock.erro).toHaveBeenCalledWith('Erro inesperado ao comunicar com o servidor.');
  });

  it('propaga o erro original pra quem chamou', () => {
    let erroRecebido: unknown;
    http.get('/produtos').subscribe({ error: (err) => (erroRecebido = err) });

    const req = httpTesting.expectOne('/produtos');
    req.flush(
      { code: 'X', message: 'y', details: [], trace_id: 'z' },
      { status: 400, statusText: 'Bad Request' },
    );

    expect(erroRecebido).toBeTruthy();
  });
});
