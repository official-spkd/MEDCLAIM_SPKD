package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/claims"
	"medclaim-backend/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Dashboard: GET /api/dashboard
func (h *Handler) Dashboard(c *gin.Context) {
	out, err := h.svc.Dashboard(c.Request.Context(), auth.Must(c), claims.FilterDari(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Anomali: GET /api/anomali
func (h *Handler) Anomali(c *gin.Context) {
	out, err := h.svc.Anomali(c.Request.Context(), auth.Must(c), claims.FilterDari(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Financial: GET /api/financial
func (h *Handler) Financial(c *gin.Context) {
	out, err := h.svc.Financial(c.Request.Context(), auth.Must(c), claims.FilterDari(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
