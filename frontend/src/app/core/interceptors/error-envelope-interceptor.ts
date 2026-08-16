import { HttpContextToken, HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';

import { ErrorEnvelope } from '../models/error-envelope';
import { NotificacaoService } from '../../shared/notificacao/notificacao-service';

export const IGNORAR_ERRO_GLOBAL = new HttpContextToken<boolean>(() => false);

function isErrorEnvelope(valor: unknown): valor is ErrorEnvelope {
  return (
    typeof valor === 'object' &&
    valor !== null &&
    'code' in valor &&
    'message' in valor
  );
}

export const errorEnvelopeInterceptor: HttpInterceptorFn = (req, next) => {
  const notificacao = inject(NotificacaoService);

  return next(req).pipe(
    catchError((err: unknown) => {
      if (req.context.get(IGNORAR_ERRO_GLOBAL)) {
        return throwError(() => err);
      }
      if (err instanceof HttpErrorResponse) {
        if (isErrorEnvelope(err.error)) {
          notificacao.erro(err.error.message);
        } else if (err.status === 0) {
          notificacao.erro('Não foi possível conectar ao servidor.');
        } else {
          notificacao.erro('Erro inesperado ao comunicar com o servidor.');
        }
      }
      return throwError(() => err);
    }),
  );
};
