export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  categoria: string;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ProdutoInput {
  codigo: string;
  descricao: string;
  saldo: number;
  categoria: string;
}

export interface ProdutoUpdateInput {
  descricao: string;
  saldo: number;
  categoria: string;
}

export interface SugestaoProduto {
  codigo: string;
  descricao_sugerida: string;
  produtos_similares: Produto[];
}
