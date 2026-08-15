package auth

import "time"

type User struct {
	ID                       int64
	Email                    string
	SenhaHash                string
	Verificado               bool
	TokenVerificacao         *string
	TokenVerificacaoExpiraEm *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
