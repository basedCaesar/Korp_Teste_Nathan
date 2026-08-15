package faturamento

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

func RegisterRoutes(r *gin.Engine, svc *Service) {
	h := &Handler{svc: svc}
	g := r.Group("/notas")
	g.POST("", h.criar)
	g.GET("", h.listar)
	g.GET("/:id", h.buscar)
	g.DELETE("/:id", h.excluir)
	g.POST("/:id/itens", h.adicionarItem)
	g.PUT("/:id/itens/:itemId", h.atualizarItem)
	g.DELETE("/:id/itens/:itemId", h.removerItem)
	g.POST("/:id/imprimir", h.imprimir)
}

func (h *Handler) criar(c *gin.Context) {
	nota, err := h.svc.Criar(c.Request.Context())
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, nota)
}

func (h *Handler) listar(c *gin.Context) {
	notas, err := h.svc.Listar(c.Request.Context())
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, notas)
}

func (h *Handler) buscar(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	nota, err := h.svc.Buscar(c.Request.Context(), id)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, nota)
}

func (h *Handler) excluir(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Excluir(c.Request.Context(), id); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) adicionarItem(c *gin.Context) {
	notaID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req AdicionarItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	item, err := h.svc.AdicionarItem(c.Request.Context(), notaID, req.ProdutoID, req.ProdutoCodigo, req.ProdutoDescricao, req.Quantidade)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) atualizarItem(c *gin.Context) {
	notaID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}
	var req AtualizarItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.RespondValidationError(c, err)
		return
	}
	item, err := h.svc.AtualizarItem(c.Request.Context(), notaID, itemID, req.Quantidade)
	if err != nil {
		h.responderErro(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) removerItem(c *gin.Context) {
	notaID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	itemID, ok := parseIDParam(c, "itemId")
	if !ok {
		return
	}
	if err := h.svc.RemoverItem(c.Request.Context(), notaID, itemID); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) imprimir(c *gin.Context) {
	notaID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if c.GetHeader("Idempotency-Key") == "" {
		httpx.RespondError(c, http.StatusBadRequest, "IDEMPOTENCY_KEY_OBRIGATORIA", "header Idempotency-Key e obrigatorio")
		return
	}
	if err := h.svc.Imprimir(c.Request.Context(), httpx.RequestID(c), notaID); err != nil {
		h.responderErro(c, err)
		return
	}
	c.Status(http.StatusOK)
}

func parseIDParam(c *gin.Context, nome string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(nome), 10, 64)
	if err != nil {
		httpx.RespondError(c, http.StatusBadRequest, "ID_INVALIDO", nome+" deve ser numerico")
		return 0, false
	}
	return id, true
}

func (h *Handler) responderErro(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotaNaoEncontrada):
		httpx.RespondError(c, http.StatusNotFound, "NOTA_NAO_ENCONTRADA", err.Error())
	case errors.Is(err, ErrItemNaoEncontrado):
		httpx.RespondError(c, http.StatusNotFound, "ITEM_NAO_ENCONTRADO", err.Error())
	case errors.Is(err, ErrNotaNaoAberta):
		httpx.RespondError(c, http.StatusConflict, "NOTA_NAO_ABERTA", err.Error())
	case errors.Is(err, ErrSaldoInsuficiente):
		httpx.RespondError(c, http.StatusConflict, "SALDO_INSUFICIENTE", err.Error())
	case errors.Is(err, ErrEstoqueIndisponivel):
		httpx.RespondError(c, http.StatusServiceUnavailable, "ESTOQUE_INDISPONIVEL", err.Error())
	default:
		slog.Error("erro inesperado no dominio faturamento", "error", err)
		httpx.RespondError(c, http.StatusInternalServerError, "ERRO_INTERNO", "erro interno do servidor")
	}
}
