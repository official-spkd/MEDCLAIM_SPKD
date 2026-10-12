// Package health: endpoint liveness dan readiness untuk orkestrator/monitoring.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger dipenuhi oleh pembungkus *sql.DB; handler tidak terikat ke driver.
type Pinger interface {
	PingContext(ctx context.Context) error
}

type Handler struct{ db Pinger }

func NewHandler(db Pinger) *Handler { return &Handler{db: db} }

// Live: proses hidup (tidak menyentuh database).
func (h *Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready: siap melayani (database terjangkau).
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.PingContext(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database tidak tersedia"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
