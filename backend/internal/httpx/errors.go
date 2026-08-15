package httpx

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErroResposta e o envelope unico de erro devolvido por todos os endpoints.
// Nunca retornamos string solta como corpo de erro.
type ErroResposta struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
	TraceID string   `json:"trace_id"`
}

// RespondError escreve o envelope de erro padrao. Handlers nao montam JSON de erro na mao,
// apenas classificam o erro (errors.Is/As) e chamam esta funcao.
func RespondError(c *gin.Context, status int, code, message string, details ...string) {
	if details == nil {
		details = []string{}
	}
	c.AbortWithStatusJSON(status, ErroResposta{
		Code:    code,
		Message: message,
		Details: details,
		TraceID: RequestID(c),
	})
}

// RespondValidationError traduz erro de binding do Gin (go-playground/validator) para o envelope.
func RespondValidationError(c *gin.Context, err error) {
	var details []string
	var verr validator.ValidationErrors
	if errors.As(err, &verr) {
		for _, fe := range verr {
			details = append(details, fe.Field()+": "+fe.Tag())
		}
	} else {
		details = []string{err.Error()}
	}
	RespondError(c, http.StatusBadRequest, "VALIDACAO", "dados invalidos", details...)
}
