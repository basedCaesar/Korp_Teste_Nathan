import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { ApplicationConfig, inject, provideAppInitializer, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import { ConfigService } from './core/config/config-service';
import { authTokenInterceptor } from './core/interceptors/auth-token-interceptor';
import { errorEnvelopeInterceptor } from './core/interceptors/error-envelope-interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideRouter(routes),
    provideAnimationsAsync(),
    provideHttpClient(withInterceptors([authTokenInterceptor, errorEnvelopeInterceptor])),
    provideAppInitializer(() => inject(ConfigService).carregar()),
  ],
};
