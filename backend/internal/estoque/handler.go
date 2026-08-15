package estoque

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"korp/internal/httpx"
)

type Handler struct {
	svc *Service
}

// RegisterRoutes monta o CRUD de produtos em /produtos, atras de JWT (cada usuario so ve os
// proprios produtos). /estoque/baixas fica fora do grupo: chamada interna faturamento->estoque,
// sem usuario logado por tras.
func RegisterRoutes(r *gin.Engine, svc *Service, idemStore *httpx.IdempotencyStore, jwtSecret string) {
	h := &Handler{svc: svc}
	g := r.Group("/produtos")
	g.Use(httpx.JWTAuth(jwtSecret))
	g.POST("", h.criar)
	g.GET("", h.listar)
	g.GET("/:id", h.buscar)
	g.PUT("/:id", h.atualizar)
	g.DELETE("/:id", h.excluir)
	g.POST("/sugestao", h.sugerir)

	r.POST("/estoque/baixas", httpx.Idempotency(idemStore), h.baixar)
}

func (h *Handler) criar(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
		return
	}
	var req CriarProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	produto, err := h.svc.Criar(c.Request.Context(), userID, req.Codigo, req.Descricao, req.Saldo)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, produto)
}

func (h *Handler) listar(c *gin.Context) {
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
		return
	}
	produtos, err := h.svc.Listar(c.Request.Context(), userID)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, produtos)
}

func (h *Handler) buscar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
		return
	}
	produto, err := h.svc.Buscar(c.Request.Context(), id, userID)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, produto)
}

func (h *Handler) atualizar(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
		return
	}
	var req AtualizarProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	produto, err := h.svc.Atualizar(c.Request.Context(), id, userID, req.Descricao, req.Saldo)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, produto)
}

func (h *Handler) excluir(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	userID, ok := httpx.UserID(c)
	if !ok {
		httpx.RespondError(c, http.StatusUnauthorized, "TOKEN_INVALIDO", "token de autenticacao invalido")
		return
	}
	if err := h.svc.Excluir(c.Request.Context(), id, userID); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) baixar(c *gin.Context) {
	var req BaixarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	itens := make([]ItemBaixa, len(req.Itens))
	for i, item := range req.Itens {
		itens[i] = ItemBaixa{ProdutoID: item.ProdutoID, Quantidade: item.Quantidade}
	}
	if err := h.svc.BaixarItens(c.Request.Context(), itens); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) sugerir(c *gin.Context) {
	var req SugestaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	resp, err := h.svc.Sugerir(c.Request.Context(), req.Codigo)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "ID_INVALIDO", "id deve ser numerico")
		return 0, false
	}
	return id, true
}

// responderErro classifica o erro de dominio (errors.Is) e delega o envelope pro httpx.
func (h *Handler) responderErro(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProdutoNaoEncontrado):
		httpx.RespondError(c, http.StatusNotFound, "PRODUTO_NAO_ENCONTRADO", err.Error())
	case errors.Is(err, ErrCodigoDuplicado):
		httpx.RespondError(c, http.StatusConflict, "CODIGO_DUPLICADO", err.Error())
	case errors.Is(err, ErrSaldoInsuficiente):
		httpx.RespondError(c, http.StatusConflict, "SALDO_INSUFICIENTE", err.Error())
	case errors.Is(err, ErrConflitoVersao):
		httpx.RespondError(c, http.StatusConflict, "CONFLITO_VERSAO", err.Error())
	case errors.Is(err, ErrIAIndisponivel):
		httpx.RespondError(c, http.StatusServiceUnavailable, "IA_INDISPONIVEL", err.Error())
	default:
		slog.Error("erro inesperado no dominio estoque", "error", err)
		httpx.RespondError(c, http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor")
	}
}
