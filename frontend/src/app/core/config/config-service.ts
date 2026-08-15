import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AppConfig } from './app-config';

@Injectable({ providedIn: 'root' })
export class ConfigService {
  private readonly http = inject(HttpClient);
  private config: AppConfig | null = null;

  async carregar(): Promise<void> {
    this.config = await firstValueFrom(this.http.get<AppConfig>('/config.json'));
  }

  get(): AppConfig {
    if (!this.config) {
      throw new Error('ConfigService.carregar() precisa rodar antes de usar get()');
    }
    return this.config;
  }
}
