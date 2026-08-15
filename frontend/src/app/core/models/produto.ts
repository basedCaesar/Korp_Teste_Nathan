export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ProdutoInput {
  codigo: string;
  descricao: string;
  saldo: number;
}

export interface ProdutoUpdateInput {
  descricao: string;
  saldo: number;
}

export interface SugestaoProduto {
  codigo: string;
  descricao_sugerida: string;
  produtos_similares: Produto[];
}
