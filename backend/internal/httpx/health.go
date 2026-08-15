package httpx

import (
	"net/http"

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
