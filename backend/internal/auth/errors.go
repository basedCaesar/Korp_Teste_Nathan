package auth

import "errors"

var (
	ErrEmailJaCadastrado  = errors.New("email ja cadastrado")
	ErrCredenciaisInvalidas = errors.New("credenciais invalidas")
	ErrEmailNaoVerificado = errors.New("email nao verificado")
	ErrTokenInvalido      = errors.New("token invalido")
	ErrTokenExpirado      = errors.New("token expirado")
)
