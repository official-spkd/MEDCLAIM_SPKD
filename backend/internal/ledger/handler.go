package ledger

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List: GET /api/ledger
func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context(), auth.Must(c), c.Query("rs"), c.Query("status"), c.Query("q"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Detail: GET /api/ledger/:id
func (h *Handler) Detail(c *gin.Context) {
	out, err := h.svc.Detail(c.Request.Context(), auth.Must(c), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Buat: POST /api/ledger
func (h *Handler) Buat(c *gin.Context) {
	var body struct {
		EpisodeID string `json:"episodeId"`
	}
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	out, err := h.svc.Buat(c.Request.Context(), auth.Must(c), body.EpisodeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Transisi: POST /api/ledger/:id/transisi
func (h *Handler) Transisi(c *gin.Context) {
	var body BodyTransisi
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	out, err := h.svc.Transisi(c.Request.Context(), auth.Must(c), c.Param("id"), body)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
