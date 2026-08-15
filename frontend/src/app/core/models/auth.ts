export interface CadastroInput {
  email: string;
  senha: string;
}

export interface CadastroResponse {
  id: number;
  email: string;
}

export interface LoginInput {
  email: string;
  senha: string;
}

export interface LoginResponse {
  token: string;
}

export interface Usuario {
  user_id: number;
  email: string;
}
