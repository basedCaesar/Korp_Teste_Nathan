export type StatusNota = 'ABERTA' | 'PROCESSANDO' | 'FECHADA';

export interface ItemNota {
  id: number;
  nota_id: number;
  produto_id: number;
  produto_codigo: string;
  produto_descricao: string;
  quantidade: number;
  created_at: string;
}

export interface Nota {
  id: number;
  numero: number;
  status: StatusNota;
  itens?: ItemNota[];
  created_at: string;
  updated_at: string;
}

export interface ItemInput {
  produto_id: number;
  produto_codigo: string;
  produto_descricao: string;
  quantidade: number;
}

export interface ItemUpdateInput {
  quantidade: number;
}
