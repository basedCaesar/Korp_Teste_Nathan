package httpx

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery substitui gin.Recovery(): recupera panics mas devolve o envelope de
// erro padrao em vez do texto puro que o Gin gera por padrao.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic recuperado", "error", recovered, "path", c.Request.URL.Path)
		RespondError(c, http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor")
	})
}
