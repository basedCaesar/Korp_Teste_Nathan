import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AuthService } from '../../features/auth/auth-service';
import { authTokenInterceptor } from './auth-token-interceptor';

describe('authTokenInterceptor', () => {
  let http: HttpClient;
  let httpTesting: HttpTestingController;
  let authMock: { token: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    authMock = { token: vi.fn() };

    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(withInterceptors([authTokenInterceptor])),
        provideHttpClientTesting(),
        { provide: AuthService, useValue: authMock },
      ],
    });

    http = TestBed.inject(HttpClient);
    httpTesting = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpTesting.verify();
  });

  it('nao adiciona Authorization quando nao ha token', () => {
    authMock.token.mockReturnValue(null);

    http.get('/qualquer').subscribe();

    const req = httpTesting.expectOne('/qualquer');
    expect(req.request.headers.has('Authorization')).toBe(false);
    req.flush({});
  });

  it('adiciona Authorization: Bearer <token> quando logado', () => {
    authMock.token.mockReturnValue('token-abc');

    http.get('/qualquer').subscribe();

    const req = httpTesting.expectOne('/qualquer');
    expect(req.request.headers.get('Authorization')).toBe('Bearer token-abc');
    req.flush({});
  });
});
