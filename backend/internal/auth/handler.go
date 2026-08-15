package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"korp/internal/httpx"
)

type Handler struct {
	svc       *Service
	jwtSecret string
}

func RegisterRoutes(r *gin.Engine, svc *Service, jwtSecret string) {
	h := &Handler{svc: svc, jwtSecret: jwtSecret}
	g := r.Group("/auth")
	g.POST("/cadastro", h.cadastrar)
	g.GET("/verificar", h.verificar)
	g.POST("/login", h.login)
	g.GET("/me", httpx.JWTAuth(jwtSecret), h.me)
}

func (h *Handler) cadastrar(c *gin.Context) {
	var req CadastroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	user, err := h.svc.Cadastrar(c.Request.Context(), req.Email, req.Senha)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": user.ID, "email": user.Email})
}

func (h *Handler) verificar(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		httpx.RespondError(c, http.StatusBadRequest, "VALIDACAO", "parametro token e obrigatorio")
		return
	}
	if err := h.svc.VerificarEmail(c.Request.Context(), token); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Senha)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) me(c *gin.Context) {
	claims, _ := httpx.GetClaims(c)
	c.JSON(http.StatusOK, gin.H{"user_id": claims["user_id"], "email": claims["email"]})
}

func (h *Handler) responderErro(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrEmailJaCadastrado):
		httpx.RespondError(c, http.StatusConflict, "EMAIL_JA_CADASTRADO", err.Error())
	case errors.Is(err, ErrCredenciaisInvalidas):
		httpx.RespondError(c, http.StatusUnauthorized, "CREDENCIAIS_INVALIDAS", err.Error())
	case errors.Is(err, ErrEmailNaoVerificado):
		httpx.RespondError(c, http.StatusForbidden, "EMAIL_NAO_VERIFICADO", err.Error())
	case errors.Is(err, ErrTokenInvalido):
		httpx.RespondError(c, http.StatusBadRequest, "TOKEN_INVALIDO", err.Error())
	case errors.Is(err, ErrTokenExpirado):
		httpx.RespondError(c, http.StatusBadRequest, "TOKEN_EXPIRADO", err.Error())
	default:
		slog.Error("erro inesperado no dominio auth", "error", err)
		httpx.RespondError(c, http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor")
	}
}
