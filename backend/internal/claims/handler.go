package claims

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"medclaim-backend/internal/auth"
	"medclaim-backend/internal/response"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// FilterDari membaca filter standar dari query string.
func FilterDari(c *gin.Context) Filter {
	return Filter{
		Rs: c.Query("rs"), Upload: c.Query("upload"), Tipe: c.Query("tipe"),
		Dari: c.Query("dari"), Sampai: c.Query("sampai"), KelasRugi: c.Query("kelasRugi"),
		Kelas: c.Query("kelas"), Sev: c.Query("sev"), Q: c.Query("q"),
	}
}

// List: GET /api/episodes
func (h *Handler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context(), auth.Must(c), FilterDari(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// Detail: GET /api/episodes/:id
func (h *Handler) Detail(c *gin.Context) {
	out, err := h.svc.Detail(c.Request.Context(), auth.Must(c), c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}
