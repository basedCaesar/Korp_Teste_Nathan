import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { of } from 'rxjs';

import { Nota } from '../../../core/models/nota';
import { NotaService } from '../nota-service';
import { NotasLista } from './notas-lista';

describe('NotasLista', () => {
  const notas: Nota[] = [
    { id: 1, numero: 1, status: 'ABERTA', created_at: '', updated_at: '' },
    { id: 2, numero: 2, status: 'FECHADA', created_at: '', updated_at: '' },
  ];

  let notaServiceMock: { listar: ReturnType<typeof vi.fn>; criar: ReturnType<typeof vi.fn> };
  let router: { navigate: ReturnType<typeof vi.fn> };

  beforeEach(() => {
    notaServiceMock = {
      listar: vi.fn().mockReturnValue(of(notas)),
      criar: vi.fn(),
    };
    router = { navigate: vi.fn() };

    TestBed.configureTestingModule({
      imports: [NotasLista],
      providers: [
        { provide: NotaService, useValue: notaServiceMock },
        { provide: Router, useValue: router },
      ],
    });
  });

  it('carrega a lista de notas no ngOnInit', () => {
    const fixture = TestBed.createComponent(NotasLista);
    fixture.detectChanges();

    expect(notaServiceMock.listar).toHaveBeenCalled();
    expect(fixture.componentInstance['notas']()).toEqual(notas);
  });

  it('cria nota nova e navega pro detalhe dela', () => {
    notaServiceMock.criar.mockReturnValue(of({ id: 9, numero: 9, status: 'ABERTA' }));
    const fixture = TestBed.createComponent(NotasLista);
    fixture.detectChanges();

    fixture.componentInstance['nova']();

    expect(router.navigate).toHaveBeenCalledWith(['/notas', 9]);
  });

  it('abrir navega pro detalhe da nota clicada', () => {
    const fixture = TestBed.createComponent(NotasLista);
    fixture.detectChanges();

    fixture.componentInstance['abrir'](notas[1]);

    expect(router.navigate).toHaveBeenCalledWith(['/notas', 2]);
  });
});
