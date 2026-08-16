package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterHealth adiciona GET /health, usado pelo healthcheck do Docker Compose.
func RegisterHealth(r *gin.Engine, service string) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": service,
		})
	})
}

// RegisterHealthDependencias adiciona GET /health/dependencias, usado pelo frontend pra
// desabilitar acoes proativamente quando uma dependencia externa esta fora do ar. Sempre
// responde 200: o status de cada dependencia vai no corpo, nao no codigo HTTP.
func RegisterHealthDependencias(r *gin.Engine, service string, dependencias map[string]func(context.Context) bool) {
	r.GET("/health/dependencias", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status := make(map[string]bool, len(dependencias))
		for nome, checar := range dependencias {
			status[nome] = checar(ctx)
		}

		c.JSON(http.StatusOK, gin.H{
			"service":      service,
			"dependencias": status,
		})
	})
}
