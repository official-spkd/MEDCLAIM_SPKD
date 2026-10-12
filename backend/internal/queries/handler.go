package queries

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List: GET /api/queries
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), auth.Must(c), c.Query("status"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Buat: POST /api/queries
func (h *Handler) Buat(c *gin.Context) {
	var body BodyBuat
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	out, err := h.svc.Buat(c.Request.Context(), auth.Must(c), body)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Jawab: POST /api/queries/:id/answer
func (h *Handler) Jawab(c *gin.Context) {
	var body struct {
		Jawaban string `json:"jawaban"`
	}
	if err := response.Bind(c, &body); err != nil {
		response.Fail(c, err)
		return
	}
	out, err := h.svc.Jawab(c.Request.Context(), auth.Must(c), c.Param("id"), body.Jawaban)
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Templates: GET /api/templates
func (h *Handler) Templates(c *gin.Context) {
	items, err := h.svc.Templates(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
