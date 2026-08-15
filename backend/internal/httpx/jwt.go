package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const ctxKeyClaims = "jwt_claims"

func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		partes := strings.SplitN(header, " ", 2)
		if len(partes) != 2 || partes[0] != "Bearer" {
			RespondError(c, http.StatusUnauthorized, "TOKEN_AUSENTE", "token de autenticacao ausente")
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(partes[1], claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
			return
		}

		c.Set(ctxKeyClaims, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (jwt.MapClaims, bool) {
	v, ok := c.Get(ctxKeyClaims)
	if !ok {
		return nil, false
	}
	claims, ok := v.(jwt.MapClaims)
	return claims, ok
}
