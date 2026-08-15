package httpx

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID e o header propagado entre servicos e usado em todo log.
const HeaderRequestID = "X-Request-Id"

const ctxKeyRequestID = "request_id"

// RequestIDMiddleware usa o X-Request-Id recebido, ou gera um novo se ausente,
// devolvendo sempre o mesmo header na resposta.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}
		c.Set(ctxKeyRequestID, id)
		c.Writer.Header().Set(HeaderRequestID, id)
		c.Next()
	}
}

// RequestID le o id da requisicao atual, setado por RequestIDMiddleware.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
